# Dodger

Dodger is a visual twitter client written with go, phantom.js and reveal.js. It is named after Terry Pratchett's [book](http://www.terrypratchett.co.uk/index.php/us/books/dodger).

Twitter timelines are a valuable source of information but they contain very large volumes of information. Suppose you follow 500 users. If you devote just half a minute a day to each one then you need 4 hours daily just to read the timeline and this does not include the time to visit the URLs they post. This is too much time. Moreover many tweets are like _cool_ or _check it_ or _give it a try_ which does not give even a small hint of the URL content so you have to visit it. This needs of course additional time. What is the most effective way to handle such a volume of information? Of course using a computer to do the job!

Dodger reads your timeline periodically and extracts the URLs from each tweet. If a tweet does not contain a URL then it is rejected. The best way to follow such realtime conversations is with a browser or another twitter client. Dodger cares only about useful URLs posted in tweets. Dodger first filters out the URLs and renders them in PNG images using [phantom.js](http://phantomjs.org). Then it creates a [reveal.js](http://lab.hakim.se/reveal-js) presentation that combines these images and their tweet text. It is much faster to browse a timeline by seeing the screenshots of the URLs rather than reading tweets and visiting URLs with the browser. Check the [demo](http://goo.gl/qM32o1) to get an idea of how dodger works and saves time. 

# Installation

This version of Dodger is a standalone application. Hopefully the next version will be a web service.

## Prerequisites

1. Install [go](http://www.golang.org)
2. Install [reveal.js](http://lab.hakim.se/reveal-js) The simplest way it to download a .zip and unzip it in the dodger directory
3. Install [phantom.js](http://phantomjs.org) Again download a .zip, unzip it and put the executable in PATH

## Building dodger

1. `git clone https://github.com/anastasop/dodger`
2. `cd dodger`
3. `go build`

# Running dodger

This version is a console application. To run it locally first you must create a twitter application for it. Visit [developer.twitter.com](http://dev.twitter.com), create a single user REST API application and configure the access tokens for it. Then create a file named `credentials.json` like the following. It must be in the same directory as the dodger executable.

```JSON
{
	"applicationOAuthToken": "NviYKVgeO",
	"applicationOAuthSecret": "a1qHe2",

	"userOAuthToken": "39051390-rq",
	"userOAuthSecret": "f6jWndV"
}
```

The runtime flags are

* `-h` display an overview of the runtime flags
* `-i` twitter timeline refresh period. Default is 1 hour
* `-p` http port. Default is 8080
* `-r` reveal.js installation directory
* `-s` maximum number of slides displayed. Default is 50
* `-e` include tweets for which rendering failed. These pages usually have flash or other plugins or are too much javascript

There is the `run.sh` script for convenience. Run this script and point your browser at `localhost:8080`. Use the arrow keys to navigate. Left, Right move thought the presentation, Down shows the actual tweet and `g` opens a browser tab with the URL. 

# TODO

1. deploy it as a web service
2. refreshing using `http-equiv` or manual is too old fashioned. Use websockets for a more interactive experience

# Bugs

1. it runs phantom.js as a subprocess wrapped with `timeout`. Implement a timeout mechanism in go so that it can run on on-unix systems.
2. If phantom.js rendering silently then you just see an empty image or a black bar. Somehow check the rendering result and discard it if necessary

Enjoy!!
 
