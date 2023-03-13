# AI Tech Blogger

This repo contains the code for generating an AI tech blogger.  This code is intended to run as a function in GCP which is triggered by a cloud event every day at 1AM EST.  It also supports being seeded with a particular title if it is included in the message data. This project is very rough at the moment and I have many more plans for adding features and stability improvements but it is currently functional and you can see the output of the AI at [www.theaitechblog.io](https://theaitechblog.io)

## Overview

The code is deployed into GCPs cloud functions and listening to a pub/sub queue that is set to have an empty message pushed into it every night using googles cloud scheduler.  When this message is received the code shoots an API request out to OpenAIs ChatGPT endpoint asking it to write a blog post in markdown.  After the output is received some basic regex checking is done to make sure it's in the correct format for posting before pulling the output repository which is setup as a jekyl blog.  Once the pull is done it puts the generated content into a new post file and pushes the new code up to the repo where the automated github pages pipeline takes over and pushes the content into production.

## Future Improvements

- [ ] Better timestamps on posts
- [ ] Better format validation (Missing spaces for headers)
- [ ] Feeding previous posts in as messages into the ChatGPT request
- [ ] Form to request content from the AI
