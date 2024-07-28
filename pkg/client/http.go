package client

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	gReddit "github.com/cameronstanley/go-reddit"
	"github.com/gabriel-vasile/mimetype"
	"github.com/go-shiori/go-readability"
	"github.com/tiendc/go-linkpreview"
)

type fileType int

const (
	unknown fileType = iota
	image
	video
)

func knownTypes(m *mimetype.MIME) fileType {
	if strings.HasPrefix(m.String(), "image") {
		return image
	} else if strings.HasPrefix(m.String(), "video") {
		return video
	}
	return unknown
}

func getMimeType(client *http.Client, url string) (*mimetype.MIME, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	mime, err := mimetype.DetectReader(resp.Body)
	if err != nil {
		return nil, err
	}

	return mime, nil
}

func cleanupUrl(url string) (string, error) {
	if strings.Contains(url, "imgur") && strings.HasSuffix(url, "gifv") {
		return strings.ReplaceAll(url, "gifv", "webm"), nil
	}

	return url, nil
}

func fixAmp(url string) string {
	return strings.Replace(url, "&amp;", "&", -1)
}

func imgElement(media gReddit.MediaMetadata) string {
	if media.S.Gif != "" {
		return fmt.Sprintf("<img src=\"%s\" /><br/>", fixAmp(media.S.Gif))
	} else if media.S.U != "" {
		return fmt.Sprintf("<img src=\"%s\" /><br/>", fixAmp(media.S.U))
	} else {
		return ""
	}
}

var ErrVideoMissingFromJSON = errors.New("video missing from json")

func GetArticle(client *RedditClient, link *gReddit.Link) (*string, error) {
	u := link.URL
	str := ""

	if link.Selftext != "" {
		str += html.UnescapeString(link.SelftextHTML)

		// replace preview.redd.it links with images
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(str))
		if err != nil {
			return nil, err
		}

		doc.Find("a[href^='https://preview.redd.it']").Each(func(_ int, s *goquery.Selection) {
			href, _ := s.Attr("href")
			s.ReplaceWithHtml(fmt.Sprintf("<img src=\"%s\" />", href))
		})

		str, _ = doc.Html()

		if link.IsSelf {
			return &str, nil
		}
	}

	if link.Media.Oembed.Type == "video" && link.Media.Oembed.HTML != "" {
		str += html.UnescapeString(link.Media.Oembed.HTML)
		re := regexp.MustCompile(`(width|height)="[^"]*"`)
		str = re.ReplaceAllString(str, "")

		return &str, nil
	}

	if len(link.MediaMetadata) > 0 {
		var b strings.Builder
		b.WriteString("<div>")
		if len(link.GalleryData.Items) > 0 {
			for _, item := range link.GalleryData.Items {
				b.WriteString(imgElement(link.MediaMetadata[item.MediaID]))
			}
		} else {
			for _, media := range link.MediaMetadata {
				b.WriteString(imgElement(media))
			}
		}
		b.WriteString("</div>")
		str += b.String()
		return &str, nil
	}

	// todo clean up
	if strings.Contains(u, "gfycat") {
		res, err := client.HttpClient.Get(u)
		if err != nil {
			return nil, err
		}

		defer res.Body.Close()
		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			return nil, err
		}

		img, _ := doc.Find("meta[property=\"og:image\"][content$=\".jpg\"]").Attr("content")
		vid, _ := doc.Find("meta[property=\"og:video:iframe\"]").Attr("content")
		width, _ := doc.Find("meta[property=\"og:video:width\"]").Attr("content")
		height, _ := doc.Find("meta[property=\"og:video:height\"]").Attr("content")

		// I don't know whether width and height is necessary with gfycat here. I'm gonna leave it here for now.
		str = fmt.Sprintf("<div><iframe src=\"%s\" width=\"%s\" height=\"%s\"/> <img src=\"%s\" class=\"webfeedsFeaturedVisual\"/></div>", vid, width, height, img)
		return &str, nil
	}

	if strings.Contains(u, "v.redd.it") {
		video := link.SecureMedia.RedditVideo
		if video == nil {
			if len(link.CrossPostParentList) == 0 {
				return nil, ErrVideoMissingFromJSON
			}
			parent := link.CrossPostParentList[0]
			video = parent.SecureMedia.RedditVideo
		}
		if video == nil {
			return nil, ErrVideoMissingFromJSON
		}
		str += fmt.Sprintf("<iframe src=\"%s\" style=\"border:none;\" /> <img src=\"%s\" class=\"webfeedsFeaturedVisual\"/>", video.FallbackURL, link.Thumbnail)
		return &str, nil
	}

	url, err := cleanupUrl(u)
	if err != nil {
		return nil, err
	}

	t, err := getMimeType(client.HttpClient, url)
	if err != nil {
		return nil, err
	}

	switch knownTypes(t) {
	case image:
		str += fmt.Sprintf("<img src=\"%s\" class=\"webfeedsFeaturedVisual \"/>", url)
		return &str, nil
	case video:
		str += fmt.Sprintf("<video><source src=\"%s\" type=\"%s\" /></video>", url, t.String())
		return &str, nil
	}

	res, err := linkpreview.Parse(url,
		linkpreview.ReturnMetaTags(true))
	if err != nil {
		log.Println("ERROR: Something went wrong while we are processing a post. ", err)
		log.Println("Reference: ", url)
		return nil, err
	}

	str += fmt.Sprintf(`<a href="%s" style="text-decoration:none;color:inherit">
	<div style="border:1px solid gray">
		<img src="%s" />
		<div style="border-top:1px solid gray;padding:4px">
			<span><strong>%s</strong></span><br />
			<span><small>%s</small></span>
		</div>
	</div></a>`, url, res.MetaTags[8].Content, res.Title, strings.Split(url, "?")[0])

	return &str, nil

	// There might be more stuff to take care of, for now just hope I covered everything.
}

func articleFromURL(ctx context.Context, client *http.Client, pageURL string) (readability.Article, error) {
	// Make sure URL is valid
	_, err := url.ParseRequestURI(pageURL)
	if err != nil {
		return readability.Article{}, fmt.Errorf("failed to parse URL: %v", err)
	}

	// Fetch page from URL
	req, err := http.NewRequestWithContext(ctx, "GET", pageURL, nil)
	if err != nil {
		return readability.Article{}, fmt.Errorf("failed to create req for page: %v", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		return readability.Article{}, fmt.Errorf("failed to fetch the page: %v", err)
	}
	defer resp.Body.Close()

	// Make sure content type is HTML
	cp := resp.Header.Get("Content-Type")
	if !strings.Contains(cp, "text/html") {
		return readability.Article{}, fmt.Errorf("URL is not a HTML document")
	}

	// Check if the page is readable
	var buffer bytes.Buffer
	tee := io.TeeReader(resp.Body, &buffer)

	parser := readability.NewParser()
	if !parser.Check(tee) {
		return readability.Article{}, fmt.Errorf("the page is not readable")
	}

	parsedURL, err := url.Parse(pageURL)
	if err != nil {
		return readability.Article{}, fmt.Errorf("failed to parse URL: %v", err)
	}

	return parser.Parse(&buffer, parsedURL)
}
