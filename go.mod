module github.com/jwnpoh/njcgpnewsfeed

go 1.16

replace github.com/jwnpoh/njcgpnewsfeed/db => ./db

replace github.com/jwnpoh/njcgpnewsfeed/cmd => ./cmd

require (
	github.com/jwnpoh/njcgpnewsfeed/cmd v0.0.0-00010101000000-000000000000
	github.com/jwnpoh/njcgpnewsfeed/db v0.0.0-00010101000000-000000000000
)
