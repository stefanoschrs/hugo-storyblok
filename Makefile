build:
	go build -o hugo-storyblok .
	which upx && upx hugo-storyblok