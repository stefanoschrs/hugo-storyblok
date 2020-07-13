package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

type Post struct {
	Id int
	Name string
	Slug string

	Title string
	Description string
	Date time.Time

	Draft bool
	Toc bool

	Keywords []string
	Tags []string

	Body string
}

func getPosts(token string) (posts []Post, err error)  {
	var body []byte
	if os.Getenv("DEV") != "" {
		body, err = ioutil.ReadFile("stories.json")
		if err != nil {
			return
		}
	} else {
		var res *http.Response

		res, err = http.Get(fmt.Sprintf("http://api.storyblok.com/v1/cdn/stories?token=%s&q=%d", token, time.Now().Unix()))
		if err != nil {
			return
		}
		defer res.Body.Close()

		body, err = ioutil.ReadAll(res.Body)
		if err != nil {
			return
		}
	}

	var data struct{
		Stories []struct {
			Id int
			Name string
			Slug string
			Content struct {
				Title string
				Description string
				Date string
				Draft bool
				Toc bool
				Keywords []struct{
					Text string
				}
				Tags []struct{
					Text string
				}
				Body string
			}
		}
	}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return
	}

	for _, story := range data.Stories {
		var keywords []string
		for _, keyword := range story.Content.Keywords {
			keywords = append(keywords, keyword.Text)
		}

		var tags []string
		for _, tag := range story.Content.Tags {
			tags = append(tags, tag.Text)
		}

		date, err2 := time.Parse("2006-01-02 15:04", story.Content.Date)
		if err2 != nil {
			err = err2
			return
		}

		posts = append(posts, Post{
			Id:          story.Id,
			Name:        story.Name,
			Slug:        story.Slug,
			Title:       story.Content.Title,
			Description: story.Content.Description,
			Date:        date,
			Draft:       story.Content.Draft,
			Toc:         story.Content.Toc,
			Keywords:    keywords,
			Tags:        tags,
			Body:        story.Content.Body,
		})
	}

	return
}

func processPost(post Post) (str string, err error) {
	var metaItems []string
	metaItems = append(metaItems, "---")
	metaItems = append(metaItems, fmt.Sprintf("title: \"%s\"", post.Title))
	metaItems = append(metaItems, fmt.Sprintf("slug: \"%s\"", post.Slug))
	metaItems = append(metaItems, fmt.Sprintf("date: %s", post.Date.Format(time.RFC3339)))
	metaItems = append(metaItems, fmt.Sprintf("description: \"%s\"", post.Description))
	metaItems = append(metaItems, fmt.Sprintf("keywords: [\"%s\"]", strings.Join(post.Keywords, "\",\"")))
	metaItems = append(metaItems, fmt.Sprintf("tags: [\"%s\"]", strings.Join(post.Tags, ",")))
	metaItems = append(metaItems, fmt.Sprintf("draft: %t", post.Draft))
	metaItems = append(metaItems, fmt.Sprintf("toc: %t", post.Toc))
	metaItems = append(metaItems, "---")
	str = strings.Join(metaItems, "\n") + "\n"

	str += post.Body

	return
}

func main() {
	token := flag.String("token", "", "Storyblok API token")
	outDir := flag.String("dir", "", "Hugo content path")
	flag.Parse()

	posts, err := getPosts(*token)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Found %d posts\n", len(posts))

	for _, post := range posts {
		processedPost, err := processPost(post)
		if err != nil {
			log.Fatal(err)
		}

		err = ioutil.WriteFile(path.Join(*outDir, post.Slug + ".md"), []byte(processedPost), 0644)
		if err != nil {
			log.Fatal(err)
		}
	}

	log.Println("Success!")
}
