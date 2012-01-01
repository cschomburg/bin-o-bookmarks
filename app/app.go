package app

import (
	"appengine"
	"appengine/user"
	"bookmarks"
	"fmt"
	"html"
	"http"
	"mustache"
	"os"
	"strings"
)

func init() {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/follow/", handleFollow)
	http.HandleFunc("/f/", handleFollow)
	http.HandleFunc("/create", handleCreate)
	http.HandleFunc("/c", handleCreate)
	http.HandleFunc("/export", handleExport)
}

func showWelcome(w http.ResponseWriter, r *http.Request, c appengine.Context) {
	loginURL, err := user.LoginURL(c, r.URL.String())
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}
	output(w, "welcome", map[string]interface{}{
		"loginURL": loginURL,
	});
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	u := user.Current(c)
	// Login page
	if u == nil {
		showWelcome(w, r, c)
		return
	}

	// Get passed tags for filtering
	tagString := r.FormValue("tags")
	tags := strings.Split(tagString, ",")

	// If tag "hidden" is not passed ...
	showHidden := false
	for _, tag := range(tags) {
		if tag == "hidden" {
			showHidden = true
			break
		}
	}
	// ... hide all "hidden" tags
	if !showHidden {
		tags = append(tags, "-hidden")
	}

	// Fetch bookmarks
	marks, err := bookmarks.ByTags(c, tags)
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}

	if tagString == "" {
		tagString = "all"
	}
	title := fmt.Sprintf("%s tagged with '%s'",
	                     pluralize("Bookmark", len(marks), true), tagString)

	logoutURL, err := user.LogoutURL(c, r.URL.String())
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}

	output(w, "index", map[string]interface{}{
		"user": u,
		"logoutURL": logoutURL,
		"title": title,
		"bookmarks": marks,
	});
}

func pluralize(text string, count int, prepend bool) string {
	if count != 1 {
		text += "s"
	}
	if prepend {
		text = fmt.Sprint(count) + " " + text
	}
	return text
}

func handleFollow(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	u := user.Current(c)
	if u == nil {
		showWelcome(w, r, c)
		return
	}

	// Extract query information, valid requests are:
	// /follow/[multiple,tags]/?q=search+string
	// /follow?q=[multiple,tags]+search+string
	tagString := ""
	query := ""
	urlParts := strings.SplitN(r.URL.Path, "/", 3)
	if len(urlParts) == 3 && urlParts[2] != "" {
		tagString = urlParts[2]
		query = r.FormValue("q")
	} else {
		queryParts := strings.SplitN(r.FormValue("q"), " ", 2)
		tagString = queryParts[0]
		if len(queryParts) > 1 {
			query = queryParts[1]
		}
	}

	// Fetch bookmarks with tags
	tags := strings.Split(tagString, ",")
	marks, err := bookmarks.ByTags(c, tags)
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}

	// If no bookmarks with these tags are found, use "default" tag with query
	// e.g. for search engine link
	if len(marks) == 0 {
		query = tagString + " " + query
		tagString = "default"
		marks, err = bookmarks.ByTags(c, []string{"default"})
	}
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}

	// Search query passed? Format URLs
	if query != "" {
		for i := 0; i < len(marks); i++ {
			marks[i].URL = strings.Replace(marks[i].URL, "%s", query, -1)
		}
	}

	// Navigate directly if a single bookmark was found
	if len(marks) == 1 {
		w.Header().Set("Location", marks[0].URL)
		w.WriteHeader(http.StatusFound)
		return
	}

	fullQuery := tagString
	if query != "" {
		 fullQuery += " " + query
	}

	logoutURL, err := user.LogoutURL(c, r.URL.String())
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}

	title := fmt.Sprintf("Following %s tagged with '%s' and query '%s'",
	                     pluralize("Bookmark", len(marks), true),
						 tagString, query)

	output(w, "index", map[string]interface{}{
		"user": u,
		"logoutURL": logoutURL,
		"count": len(marks),
		"title": title,
		"query": fullQuery,
		"bookmarks": marks,
	});
}

func handleCreate(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	u := user.Current(c)
	if u == nil {
		return
	}

	r.ParseForm()
	if r.Form["url"] == nil {
		output(w, "create", map[string]interface{}{
			"user": u,
		});
		return
	}

	for i, url := range(r.Form["url"]) {
		title := ""
		if len(r.Form["title"]) > i {
			title = r.Form["title"][i]
		} else {
			title = url
		}

		var tags []string
		if len(r.Form["tags"]) > i {
			tags = strings.Split(r.Form["tags"][i], ",")
		}

		bm := bookmarks.NewBookmark(u, url, title, tags)
		_, err := bm.Save(c)
		if err != nil {
			http.Error(w, err.String(), http.StatusInternalServerError)
			return
		}
	}

	handleIndex(w, r)
}

func handleExport(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	u := user.Current(c)
	if u == nil {
		return
	}

	//urlParts := strings.SplitN(r.URL.Path, ".", 2)

	marks, err := bookmarks.ByTags(c, []string{})
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}

	output(w, "export", map[string]interface{}{
		"user": u,
		"count": len(marks),
		"bookmarks": marks,
	});
}

func render(view string, context ...interface{}) string {
	return mustache.RenderFile("views/" + view + ".mustache", context...)
}

func output(w http.ResponseWriter, view string, context ...interface{}) {
	fmt.Fprintln(w, render(view, context...))
}

func error(w http.ResponseWriter, err os.Error) {
	fmt.Fprintln(w, html.EscapeString(err.String()))
}
