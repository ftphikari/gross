GRoSS - Go Reader of Simple Syndications

API:
/ - show all feeds with links to /news/hash
/?add=link - add feed with that link
/?del=hash - delete feed with that hash
/?import=file - import feeds from file
/?export=1 - export current feed list to downloadable file
/?toggleseen=1 - toggle 'Hide already seen'

/update - update all feeds

/news - show all news sorted with links to /news/hash/hash
/news/hash - show all news in hashed feed sorted with links to /news/hash/hash
/news/hash/hash - show hashed item from hashed feed

/news?from=50 - start showing news since position 50
/news/hash?from=50 - start showing feed news since position 50

UI:
Index page has buttons [add, del, import, export, toggleseen]

add - checkbox that opens url input box below with an "add" button
del - <a href="/?del=hash"> attached to every feed
import - checkbox that opens file chooser below with an "import button"
export - <a href="/?export">

FORMAT:
feeds should just be a list of urls

X when / is requested, go over all links, hash them in the process,
get feed title, get errors from map[hash]err
X when /news is requested, go over all links, hash them in the process,
collect feed items, sort them, display title, link to /news/hash/hash
* when /news?update is requested, go over all links, hash them in the process,
update each one
some of them may err, store them in the map[hash]err, which is used by / and /news/hash
map of errors is cleaned on every /news?update
