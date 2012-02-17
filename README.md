# Bin o'Bookmarks

Bin o'Bookmarks is an online bookmarks storage for Google App Engine.
To keep it simple and minimal, it provides a flexible categorization
and search based solely on tags combined with operators.
It's main reasons are to escape cross-browser syncing, to unify
multiple services (bookmarks, search engines, reading lists) and to provide a
simple extendable API.

## Features

* Tag-based navigation
* Search engine query support
* Simple API and storage scheme (unique url, title, tags)
* Bookmarklet support

## Getting Started

1. Get the code.
2. Download the [Google App Engine SDK for Go](https://code.google.com/appengine/downloads.html#Google_App_Engine_SDK_for_Go).
3. Create a [Google App Engine application](https://appengine.google.com/).
4. Change the `application` value in `app.yaml` to your app name from step 3.
5. You can run the app locally by running `dev_appserver.py`(see [Getting Started](https://code.google.com/appengine/docs/go/gettingstarted/))
6. Deploy your app to Google App Engine (see [Uploading Your Application](https://code.google.com/appengine/docs/go/gettingstarted/uploading.html))

## Instructions

* Chain multiple tags with a comma (,)
* `%s` in URLs get replaced with your search terms in Follow mode.
* Follow mode expects a form of `multiple,tags search terms` - both parts are optional. For example `blog,coding` lists all bookmarks that have both `blog` and `coding` as tags, and `google some things` would open the bookmarks with tag `google` and format the URL with "some things".
* Bookmarks tagged with `hidden` are not visible in your main listing.
* If Follow mode doesn't find any bookmarks with your tag list, it shows all bookmarks tagged as `default`
* Prefixing a tag with `-` (negate) hides its bookmarks in listings.
* Prefix a tag with `!` (unique) while creating a bookmark to remove this tag from all other bookmarks.
* Use tag `-follow`to disable automatic redirection if there was only one link found.

## Tips & Tricks

* Group your favorite websites with a tag like `favorite` or `top` and use this listing as your start page in your browser (`/?q=favorite`).
* Set the query URL (`/?q=%s`) as your default search provider in your browser to quickly navigate to all your favorite sites/searches.
* Implement a tag like `readinglist` and use a bookmarklet to save interesting articles for later reading.
* Use unique tags for bookmarks to quickly navigate to them. Especially useful as your personal small URL shortener.
* Cross-device link sharing: Use a bookmarklet to store a website in an unique tag like `!link` and a bookmark on your other device targeting `/?q=link` to quickly open it.
* Store all your search engines with a custom tag like `search` and use `%s` instead of the query in the bookmark URL. Now you can get access to all your search engines with `search my custom terms`
* Tag a single search engine as `default` to use it as a fallback in Follow mode - or multiple ones for a quick listing.

## Planned features

* Export option
* Generic bookmarklet asking for tags
* Android share intent
* Better mobile style
