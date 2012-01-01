package bookmarks

import (
	"appengine"
	"appengine/datastore"
	"appengine/user"
	"os"
	"strings"
	"url"
)

type Bookmark struct {
	UserId string
	URL string
	Title string
	Tags []string
	TimeCreated int64
	TimeUpdated int64
}

type Tag struct {
	Name string
	Bookmarks []Bookmark
}

func NewBookmark(u *user.User, url, title string, tags []string) Bookmark {
	return Bookmark{u.Id, url, title, tags, 0, 0}
}


func (b Bookmark) FaviconURL() string {
	domain := ""
	u, err := url.Parse(b.URL)
	if err == nil {
		domain = u.Host
	}
	return "http://www.google.com/s2/u/0/favicons?domain=" + domain
}

func (b Bookmark) TagList() string {
	return strings.Join(b.Tags, ",")
}

func (b *Bookmark) Save(c appengine.Context) (success bool, err os.Error) {
	if b.URL == "" {
		return false, nil
	}

	key, err := Exists(c, *b)
	if err != nil {
		return false, err
	}
	if key == nil {
		key = datastore.NewIncompleteKey(c, "Bookmark", nil)
		b.TimeCreated, _, err = os.Time()
	}
	b.TimeUpdated, _, err = os.Time()
	if err != nil {
		return false, err
	}

	_, err = datastore.Put(c, key, b)
	return err != nil, err
}

func (b *Bookmark) Delete(c appengine.Context) (success bool, err os.Error) {
	if b.URL == "" {
		return false, nil
	}

	key, err := Exists(c, *b)
	if key == nil || err != nil  {
		return false, err
	}

	err = datastore.Delete(c, key)
	return err != nil, err
}

func ByTags(c appengine.Context, tags []string) (bms []Bookmark, err os.Error) {
	q := datastore.NewQuery("Bookmark").Filter("UserId=", user.Current(c).Id).Order("Title")
	for _, tag := range(tags) {
		if tag != "" {
			q.Filter("Tags=", tag)
		}
	}
	count, err := q.Count(c)
	if err != nil {
		return
	}
	bms = make([]Bookmark, 0, count)
	_, err = q.GetAll(c, &bms)
	return bms, err
}

func Exists(c appengine.Context, b Bookmark) (key *datastore.Key, err os.Error) {
	q := datastore.NewQuery("Bookmark").Filter("UserId=", user.Current(c).Id).Filter("URL=", b.URL).KeysOnly()
	keys, err := q.GetAll(c, nil)
	switch l := len(keys); true {
	case l > 1:
		return nil, os.NewError("Multiple bookmarks in datastore found for " + b.URL)
	case l == 1:
		return keys[0], nil
	}
	return nil, nil
}

func FilterTag(bms []Bookmark, tag string) []Bookmark {
	var filtered []Bookmark
	for _, b := range bms {
		found := false
		for _, btag := range(b.Tags) {
			if btag == tag {
				found = true
				break
			}
		}
		if (!found) {
			filtered = append(filtered, b)
		}
	}
	return filtered
}
