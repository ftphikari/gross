GRoSS - Go Reader of Simple Syndications

API:

x / - show all news sorted with links to /hash/hash
x /?from=50 - start showing news since position 50

x /hash - show all news in hashed feed sorted with links to /hash/hash
x /hash?from=50 - start showing feed news since position 50
x /hash/hash - show hashed item from hashed feed

x /feeds - show all feeds with links to /hash
x /feeds?add=url - add feed
x /feeds?delete=hash - delete feed
/feeds?autoredirect=hash - toggle autoredirect for hash

/import?feed=file - import feeds from file
/export - export current feed list to downloadable file

x /refresh - refresh all feeds
/toggleread - toggle 'Hide already read'
/readall - mark everything as 'Already read'

UI:
Index page has buttons [refresh, toggleseen, readall]

x refresh - <a href="/refresh">
toggleread - <a href="/toggleread">
readall - <a href="/readall">

Feeds page has buttons [add, del, import, export, toggleseen, autoredirect]

add - checkbox that opens url input box below with an "add" button
x del - <a href="/feeds?delete=hash"> attached to every feed
import - checkbox that opens file chooser below with an "import button"
export - <a href="/export">
autoredirect - <a href="/feeds?autoredirect=hash">

FORMAT:
feeds - []{url, title, autoredirect}
