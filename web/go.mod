module github.com/jwnpoh/njcgpnewsfeed/web

go 1.16

replace github.com/jwnpoh/njcgpnewsfeed/db => ../db

require (
	github.com/google/uuid v1.1.2
	github.com/jwnpoh/njcgpnewsfeed/db v0.0.0-00010101000000-000000000000
)
