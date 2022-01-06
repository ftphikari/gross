GRoSS - Go Reader of Simple Syndications

API:

x / - show all news sorted with links to /hash/hash
x /?from=50 - start showing news since position 50

x /hash - show all news in hashed feed sorted with links to /hash/hash
x /hash?from=50 - start showing feed news since position 50
x /hash/hash - show hashed item from hashed feed

x /feeds - show all feeds with links to /hash
/feeds?add=url - add feed
/feeds?delete=hash - delete feed
/feeds?autoredirect=hash - toggle autoredirect for hash

/import?feed=file - import feeds from file
/export - export current feed list to downloadable file

x /update - update all feeds
/toggleseen - toggle 'Hide already seen'

UI:
Index page has buttons [add, del, import, export, toggleseen]

add - checkbox that opens url input box below with an "add" button
del - <a href="/?del=hash"> attached to every feed
import - checkbox that opens file chooser below with an "import button"
export - <a href="/?export">

FORMAT:
feeds - []{url, title, autoredirect}
TODO: save as OPML
