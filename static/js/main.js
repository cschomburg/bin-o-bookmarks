$(function() {
	$('#filter_bookmarks').keyup(function() {
		var keywords = $(this).val().toLowerCase().split(",");

		$('#bookmarks .bookmark').each(function(i) {
			bookmark = $(this);
			var url = bookmark.data("url").toLowerCase();
			var title = bookmark.data("title").toLowerCase();
			var tags = bookmark.data("tags").toLowerCase();

			var matches = true;
			$.each(keywords, function(i, keyword) {
				if (!keyword)
					return true;
				if (url.indexOf(keyword) != -1)
					return true;
				if (title.indexOf(keyword) != -1)
					return true;
				if (tags.indexOf(keyword) != -1)
					return true;
				matches = false;
				return false
			});
			$(this).toggle(matches)
		});
	});

	$('#bookmarks .bookmark .edit').click(function() {
		var bookmark = $(this).parents(".bookmark");
		$('#create .url').val($(bookmark).data('url'));
		$('#create .title').val($(bookmark).data('title'));
		$('#create .tags').val($(bookmark).data('tags'));
	});
});
