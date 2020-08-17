# Go-web 

Web application which takes a website URL as an input and provides general information about the contents of the page:
- HTML Version
- Page Title
- Headings count by level
- Amount of internal and external links - Amount of inaccessible links
- Amount of inaccessible links
- If a page contains a login form


## Run the application: 
Run the app with: 
```
go run main.go "some/url"
```

Run tests with:
``` 
go test
```

# Requirements
This app requires Go1.1+ 
In addition, this app uses Goquery (see go.mod file) and the net/html package. Both require UTF-8 encoding. 
Instead of using Goquery I tried the "golang.org/x/net/html" and its tokenizer but Goquery seemed more like a real work project. Traversing the DOM tree works similar.

# Notes
- The fetchResult excludes duplicates which may not be required
- I use a Matcher Interface since we talked about it, although it's probably unnecessary here. See v1 for simpler sorting
- What I would change: improve error handling, better testing
- use channels for sort result
