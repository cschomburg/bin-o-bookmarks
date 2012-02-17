/*
	bookmarks.go - backend/database code for Bin o'Bookmarks

	Copyright (C) 2012  Constantin "xConStruct" Schomburg <me@xconstruct.net>

	Bin o'Bookmarks is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	Bin o'Bookmarks is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.

	You should have received a copy of the GNU General Public License
	along with Foobar.  If not, see <http://www.gnu.org/licenses/>.
*/

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
	TimeUpdated int64
}

type Tag struct {
	Name string
	Bookmarks []Bookmark
}

func NewBookmark(u *user.User, url, title string, tags []string) Bookmark {
	return Bookmark{u.Id, url, title, tags, 0}
}


func (b Bookmark) FaviconURL() string {
	domain := ""
	u, err := url.Parse(b.URL)
	if err == nil {
		domain = u.Host
	}
	return "http://www.google.com/s2/u/0/favicons?domain=" + domain
}

func (b Bookmark) TagString() string {
	return strings.Join(b.Tags, ",")
}

func (b *Bookmark) Save(c appengine.Context) (success bool, err os.Error) {
	if b.URL == "" {
		return false, nil
	}

	if b.Title == "" {
		b.Title = b.URL
	}

	key, err := Exists(c, *b)
	if err != nil {
		return false, err
	}

	if key == nil {
		key = datastore.NewIncompleteKey(c, "Bookmark", nil)
	}
	b.TimeUpdated, _, err = os.Time()
	if err != nil {
		return false, err
	}

	// "!tag" makes this tag unique: the tag will be removed from all other
	// bookmarks in the datastore
	for i, tag := range(b.Tags) {
		if tag == "" {
			continue;
		}
		op := tag[0:1]
		if op == "!" {
			tag = tag[1:]
			b.Tags[i] = tag
			DeleteTag(c, tag)
		}
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

func DeleteTag(c appengine.Context, tag string) (err os.Error) {
	// Fetch bookmarks with this tag
	q := datastore.NewQuery("Bookmark").Filter("UserId=", user.Current(c).Id).Filter("Tags=", tag)
	count, err := q.Count(c)
	if err != nil {
		return err
	}
	var bms []Bookmark
	keys, err := q.GetAll(c, &bms)
	if err != nil {
		return err
	}

	// Remove tag from bookmark
	bmsRef := make([]interface{}, count)
	for i := 0; i < len(bms); i++ {
		bmsRef[i] = &bms[i]
		btags := bms[i].Tags
		for j := 0; j < len(btags); j++ {
			if btags[j] == tag {
				bms[i].Tags = append(btags[:j], btags[j+1:]...)
				break;
			}
		}
	}

	// Put them back on the datastore
	_, err = datastore.PutMulti(c, keys, bmsRef)
	return err
}

func ByTags(c appengine.Context, tags []string) (bms []Bookmark, err os.Error) {
	q := datastore.NewQuery("Bookmark").Filter("UserId=", user.Current(c).Id).Order("Title")

	// Build query
	var negTags []string
	for _, tag := range(tags) {
		if tag != "" {
			op := tag[0:1]
			switch op {
			case "-": negTags = append(negTags, tag[1:])
			case "!": q.Filter("Tags=", tag[1:])
			default:  q.Filter("Tags=", tag)
			}
		}
	}

	count, err := q.Count(c)
	if err != nil {
		return
	}
	bms = make([]Bookmark, 0, count)
	_, err = q.GetAll(c, &bms)

	bms = FilterTags(bms, negTags)

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

func FilterTags(bms []Bookmark, tags []string) []Bookmark {
	if len(tags) == 0 {
		return bms
	}

	var filtered []Bookmark
	for _, b := range bms {
		found := false
		BTAGS: for _, btag := range(b.Tags) {
			for _, tag := range(tags) {
				if btag == tag {
					found = true
					break BTAGS
				}
			}
		}
		if !found {
			filtered = append(filtered, b)
		}
	}
	return filtered
}

func ContainsTag(tags []string, tag string) (has bool, i int) {
	for i, t := range(tags) {
		if t == tag {
			return true, i
		}
	}
	return false, 0
}
