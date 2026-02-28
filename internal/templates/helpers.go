package templates

// TagFilterClass returns the CSS class for a tag filter link (selected vs not).
func TagFilterClass(selected, current string) string {
	if selected == current {
		return "rounded-full px-2.5 py-0.5 text-sm bg-amber-100 text-amber-800"
	}
	return "rounded-full px-2.5 py-0.5 text-sm bg-stone-100 text-stone-600 hover:bg-stone-200"
}
