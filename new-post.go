package aiblog

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/go-git/go-billy/v5/memfs"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/go-git/go-git/v5/storage/memory"
	_ "github.com/joho/godotenv/autoload"
	"github.com/sashabaranov/go-openai"
)

var client *openai.Client
var openaiToken string
var githubUser string
var githubToken string
var githubRepo string

var getTitleRegex = regexp.MustCompile(`(?m)^title:\s+(.*)$`)
var contentForm = regexp.MustCompile(`(?m)---\s(?:.*\s)+---[\s\S]*`)

const gptNewPostPrompt = `Act as a tech blogger.  Write a blog post using markdown syntax about a current relevant topic in programming. At the top of the post write a metadata block in this syntax
---
title: %s 
categories: []
tags: []
---
`

func init() {
	// Fetch end variables
	openaiToken = os.Getenv("OPENAI_API_TOKEN")
	githubUser = os.Getenv("GITHUB_USER")
	githubToken = os.Getenv("GITHUB_TOKEN")
	githubRepo = os.Getenv("GITHUB_REPO")

	// TODO Validate env is not null

	// Create openai client
	client = openai.NewClient(os.Getenv("OPENAI_API_TOKEN"))

	// Register Function
	functions.CloudEvent("CreateNewPost", createPost)
}

// TODO integrate previous posts into messages (Store in DB?, persistent message queue?)
func generateBlogPostContent(title string) (string, error) {
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: fmt.Sprintf(gptNewPostPrompt, title),
				},
			},
		},
	)

	if err != nil {
		return "", err
	}

    content := contentForm.FindString(resp.Choices[0].Message.Content)
    if content == "" {
        return "", errors.New("content generated in wrong form")
    }

	return content, nil
}

func generateBlogPostFileName(content string) string {
	title := getTitleRegex.FindStringSubmatch(content)[1]
	formattedTitle := strings.ReplaceAll(strings.ToLower(title), " ", "-")
	return time.Now().Format(fmt.Sprintf("2006-01-02-%s.md", formattedTitle))
}

func commitPost(content string) error {
	storage := memory.NewStorage()
	fs := memfs.New()

	auth := &http.BasicAuth{
		Username: githubUser,
		Password: githubToken,
	}

	r, err := git.Clone(storage, fs, &git.CloneOptions{
		URL: githubRepo,
	})
	if err != nil {
		return err
	}

	w, err := r.Worktree()
	if err != nil {
		return err
	}

	filePath := filepath.Join("_posts", generateBlogPostFileName(content))
	newPost, err := fs.Create(filePath)
	if err != nil {
		return err
	}

	newPost.Write([]byte(content))
	newPost.Close()

	w.Add(filePath)
	w.Commit(fmt.Sprintf("Added new post - %s", getTitleRegex.FindStringSubmatch(content)[1]), &git.CommitOptions{
		Author: &object.Signature{
			Email: "graytonio.ward@gmail.com",
			Name:  "AI Author",
			When:  time.Now(),
		},
	})

	return r.Push(&git.PushOptions{
		RemoteName: "origin",
		Auth:       auth,
	})
}

type PubSubMessage struct {
    Message struct {
        Data []byte `json:"data"`
    } `json:"message"`
}

func createPost(ctx context.Context, e event.Event) error {
    log.Println("Parsing Event Data")
    var m PubSubMessage
    err := e.DataAs(&m)
    if err != nil {
        return err
    }	
    log.Printf("Requested Title: %s", string(m.Message.Data))
	
    log.Println("Generating Post Content")
	content, err := generateBlogPostContent(string(m.Message.Data))
	if err != nil {
		return err
	}

	log.Println("Pushing New Post to Github")
	return commitPost(content)
}
