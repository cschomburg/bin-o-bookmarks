package app

import (
	"appengine"
	"appengine/user"
	"bookmarks"
	"fmt"
	"http"
	"mustache"
	"strings"
)

func init() {
	http.HandleFunc("/", handleIndex)
	http.HandleFunc("/welcome", handleWelcome)
	http.HandleFunc("/create", handleCreate)
	http.HandleFunc("/c", handleCreate)
	http.HandleFunc("/export", handleExport)
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

func handleWelcome(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r);
	output(c, w, "welcome");
}

func handleIndex(w http.ResponseWriter, r *http.Request) {
	c := appengine.NewContext(r)
	u := user.Current(c)
	if u == nil {
		output(c, w, "welcome");
		return
	}

	// Extract query information in form of ?q=multiple,tags+search+string
	fullQuery := r.FormValue("q")
	queryParts := strings.SplitN(fullQuery, " ", 2)
	tagString := queryParts[0]
	query := ""
	if len(queryParts) > 1 {
		query = queryParts[1]
	}

	// Get tag array (from tagString or default)
	var tags []string
	if tagString == "" && query == "" {
		tags = []string{"-follow"}
		fmt.Println("somehow")
	} else {
		tags = strings.Split(tagString, ",")
	}

	// Follow mode or just listing?
	followMode := true
	if has, i := bookmarks.ContainsTag(tags, "-follow"); has {
		followMode = false
		tags = append(tags[:i], tags[i+1:]...)
	}

	// If tag "hidden" is not passed, hide all "hidden" tags
	if has, _ := bookmarks.ContainsTag(tags, "hidden"); !has {
		tags = append(tags, "-hidden")
	}

	// Fetch bookmarks with tags
	marks, err := bookmarks.ByTags(c, tags)
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}

	// If no bookmarks with these tags are found, use "default" tag with query
	// as fallback, e.g. for search engine link
	if len(marks) == 0 {
		query = fullQuery
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
	if followMode && len(marks) == 1 {
		w.Header().Set("Location", marks[0].URL)
		w.WriteHeader(http.StatusFound)
		return
	}

	// Create title
	title := pluralize("Bookmark", len(marks), true)
	if tagString != "" {
		title += " tagged with '" + tagString + "'"
	}
	if query != "" {
		if tagString == "" {
			title += " with"
		} else {
			title += " and"
		}
		title += " query '" + query + "'"
	}

	output(c, w, "index", map[string]interface{}{
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
		output(c, w, "create");
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

		// "!tag" makes this tag unique: the tag will be removed from all other
		// bookmarks in the datastore
		for i, tag := range(tags) {
			if tag == "" {
				continue;
			}
			op := tag[0:1]
			if op == "!" {
				tag = tag[1:]
				tags[i] = tag
				bookmarks.DeleteTag(c, tag)
			}
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

	marks, err := bookmarks.ByTags(c, []string{})
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}

	output(c, w, "export", map[string]interface{}{
		"count": len(marks),
		"bookmarks": marks,
	});
}

func render(view string, context ...interface{}) string {
	return mustache.RenderFile("views/" + view + ".mustache", context...)
}

func output(c appengine.Context, w http.ResponseWriter, view string, context ...interface{}) {
	// Get user info
	u := user.Current(c)
	loginURL, err := user.LoginURL(c, appengine.DefaultVersionHostname(c))
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}
	logoutURL, err := user.LogoutURL(c, appengine.DefaultVersionHostname(c))
	if err != nil {
		http.Error(w, err.String(), http.StatusInternalServerError)
		return
	}

	context = append(context, map[string]interface{}{
		"user": u,
		"loginURL": loginURL,
		"logoutURL": logoutURL,
	})
	fmt.Fprintln(w, render(view, context...))
}
