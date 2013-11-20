
package main

import (
	"container/list"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/go-oauth/oauth"
	"github.com/golang/glog"
)


type Tweet struct {
	Id int64
	Text string
	User struct {
		Id int64
		Name string
		Screen_name string
	}
	Entities struct {
		Urls []struct {
			Url string
			Expanded_url string
		}
	}
}

type Slide struct {
	UserName string
	UserScreenName string
	Text string
	Expanded_url string
	Base64PNGencoding string
}

type Presentation struct {
	RefreshSeconds int
	Slides []*Slide
}

// flags
var (
	reveal_loc = flag.String("r", "", "reveal.js installation directory")
	refresh_duration = flag.Int("i", 60, "refresh internal in minutes")
	show_non_rendered = flag.Bool("e", true, "show tweets for which rendering failed")
	max_slides = flag.Int("s", 50, "maximum number of slides displayed")
	http_port = flag.Int("p", 8080, "http port")
)

// configuration
var (
	user_credentials oauth.Credentials
	oauth_client oauth.Client
	ignore_hosts []string
)

var (
	present = list.New()
)


func slides_handler(w http.ResponseWriter, r *http.Request) {
	// this reparses the template for every request
	// useful for development as you can change the presentation
	// without stopping the server
	slides_template := template.Must(template.New("slides").ParseFiles("./slides.html"))
	s := make([]*Slide, 0, present.Len())
	for e := present.Front(); e != nil; e = e.Next() {
		s = append(s, e.Value.(*Slide))
	}
	w.Header().Set("Cache-Control", "no-cache")
	// server refreshes timeline every refresh_duration minutes
	// browser refreshed 10 minutes afterwards. 10 minutes should be enough
	// to render all urls
	slides_template.Execute(w, Presentation{(*refresh_duration + 10) * 60, s})
}


func main() {
	flag.Parse()

	if *reveal_loc == "" {
		glog.Error("must specify reveal.js directory with -r")
		os.Exit(2)
	}

	if b, err := ioutil.ReadFile("./ignoreHosts.json"); err == nil {
		if err := json.Unmarshal(b, &ignore_hosts); err != nil {
			glog.Fatal(err)
		}
	} else {
		glog.Fatal(err)
	}

	creds_json, err := ioutil.ReadFile("./credentials.json")
	if err != nil {
		glog.Error(err)
		os.Exit(2)
	}
	creds_map := make(map[string]string)
	err = json.Unmarshal(creds_json, &creds_map)
	if err != nil {
		glog.Fatal(err)
	}
	user_credentials.Token = creds_map["userOAuthToken"]
	user_credentials.Secret = creds_map["userOAuthSecret"]
	oauth_client = oauth.Client {
		oauth.Credentials {
			Token: creds_map["applicationOAuthToken"],
			Secret: creds_map["applicationOAuthSecret"],
		},
		// not used since we are already have a token
		// but included for reference
		"https://api.twitter.com/oauth/request_token",
		"https://api.twitter.com/oauth/authorize",
		"https://api.twitter.com/oauth/access_token",
	}

	glog.Info("starting background renderer")
	go updateTimeline(time.Tick(time.Duration(*refresh_duration) * time.Minute))
	
	glog.Info("starting http server")
	for _, dir := range []string{ "/css/", "/js/", "/images/", "/lib/", "/plugin/" } {
		http.Handle(dir, http.StripPrefix(dir, http.FileServer(http.Dir(path.Join(*reveal_loc, dir)))))
	}
	http.HandleFunc("/", slides_handler)
	glog.Fatal(http.ListenAndServe(":" + strconv.Itoa(*http_port), nil))
}

func updateTimeline(doUpdate <-chan time.Time) {
	// "working with timelines"
	// https://dev.twitter.com/docs/working-with-timelines

	// TODO since we don't use max_id we will lose some
	// tweets if the timeline is updated with more than
	// `count` tweets during the refresh interval

	count, since_id := int64(*max_slides), int64(1)
	for {
		glog.V(1).Info("updating timeline")
		v := url.Values{}
		v.Set("count", strconv.FormatInt(int64(count), 10))
		v.Set("since_id", strconv.FormatInt(since_id, 10))

		var timeline []Tweet
		var err error
		var resp *http.Response
		var resp_json []byte
		resp, err = oauth_client.Get(http.DefaultClient, &user_credentials, "https://api.twitter.com/1.1/statuses/home_timeline.json", v)
		if err == nil {
			if resp.StatusCode != http.StatusOK {
				err = fmt.Errorf("twitter API returned %d", resp.StatusCode)
			} else {
				if resp_json, err = ioutil.ReadAll(resp.Body); err == nil {
					if err = json.Unmarshal(resp_json, &timeline); err != nil {
						// TODO write resp_json to a file for debugging
					}
				}
			}
		}
		if err == nil {
			// not clear from docs if the tweets are ordered by Id so we scan
			since_id = timeline[0].Id
			for _, tweet := range timeline {
				if tweet.Id > since_id {
					since_id = tweet.Id
				}
			}
			glog.V(1).Info("rendering started")
			renderNewSlides(timeline)
			glog.V(1).Info("rendering finished")
		} else {
			glog.Errorln("failed to update timeline:", err)
			addSlide("dodger", "dodger", err.Error(), "", "")
		}

		// wait for next tick
		<-doUpdate
	}
}

func urlToRender(origurl string) (geturl string, ignore bool) {
	// almost all urls in tweets are from shorteners
	// so we do the redirect here
	// to check if we should ignore it
	resp, err := http.Head(origurl)
	if err != nil {
		glog.Errorln("HEAD failed:", err)
		return "", true
	}

	geturl = resp.Request.URL.String()
	u, err := url.Parse(geturl)
	if err != nil {
		glog.Infoln("malformed url:", geturl)
		return "", true
	}

	for _, h := range ignore_hosts {
		if strings.HasSuffix(strings.Split(u.Host, ":")[0], h) {
			glog.V(1).Infoln("ignoring:", geturl)
			return "", true
		}
	}
	return geturl, false
}

func addSlide(UserName, UserScreenName, Text, Expanded_url, Base64PNGencoding string) {
	present.PushFront(&Slide{UserName, UserScreenName, Text, Expanded_url, Base64PNGencoding})
	for present.Len() > *max_slides {
		present.Remove(present.Back())
	}
}

func renderNewSlides(timeline []Tweet) {
	for _, tweet := range timeline {
		for _, e := range tweet.Entities.Urls  {
			geturl, ignore := urlToRender(e.Expanded_url)
			if ignore {
				continue
			}
			glog.V(1).Infof("rendering %q", geturl)
			// running more workers concurrently would be nice
			// but since this usually runs once in an hour in the background
			// while the user is actively doing sth else, one worker is fine for now
			cmd := exec.Command("timeout", "20s", "phantomjs", "capture.js", geturl)
			body, err := cmd.Output()
			if err == nil {
				addSlide(tweet.User.Name, tweet.User.Screen_name, tweet.Text, geturl, string(body))
			} else {
				glog.Warningf("rendering %q failed: %s", geturl, err)
				if *show_non_rendered {
					addSlide(tweet.User.Name, tweet.User.Screen_name, tweet.Text, geturl, "")
				}
			}
		}
	}
}
