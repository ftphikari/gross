GRoSS - Go Reader of Simple Syndications

API:

O / - show all news sorted with links to /hash/hash
O /?from=50 - start showing news since position 50

o /hash - show all news in hashed feed sorted with links to /hash/hash
O /hash?from=50 - start showing feed news since position 50
O /hash/hash - redirect to the original link
O /hash/hash?see - mark item as 'Already seen'

O /feeds - show all feeds with links to /hash
O /feeds?add=url - add feed
O /feeds?delete=hash - delete feed

O /import?file=file - import feeds from file
O /export - export current feed list to downloadable file

O /refresh - refresh all feeds
O /toggleseen - toggle 'Hide already seen'
O /seeall - mark everything as 'Already seen'

UI:
Index page has buttons [O refresh, O export, O toggleseen, O seeall]
Feeds page has buttons [O add, O del, O import]
