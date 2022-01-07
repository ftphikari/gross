GRoSS - Go Reader of Simple Syndications

API:

x / - show all news sorted with links to /hash/hash
x /?from=50 - start showing news since position 50

x /hash - show all news in hashed feed sorted with links to /hash/hash
x /hash?from=50 - start showing feed news since position 50
O /hash/hash - redirect to the original link

x /feeds - show all feeds with links to /hash
O /feeds?add=url - add feed
O /feeds?delete=hash - delete feed

O /import?file=file - import feeds from file
O /export - export current feed list to downloadable file

O /refresh - refresh all feeds
/toggleread - toggle 'Hide already read'
/readall - mark everything as 'Already read'

UI:
Index page has buttons [O refresh, O export, toggleseen, readall]

toggleread - <a href="/toggleread">
readall - <a href="/readall">

Feeds page has buttons [O add, x del, O import, toggleseen]

x del - <a href="/feeds?delete=hash"> attached to every feed
toggleseen - ?

FORMAT:
feeds - []{url, title}
