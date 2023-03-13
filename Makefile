
deploy:
	gcloud functions deploy create-blog-post --retry --entry-point CreateNewPost --runtime go120 --trigger-topic create-ai-post --env-vars-file env-vars.yaml --gen2
