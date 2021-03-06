# universal_p2p
expose universal http front for multi p2p backend


ZeroNet/i2p/twister/freenet makes a step to p2p website, they all expose a 
http user interface,but all these model lacks well big file support 
(stream video/download iso fast enough, for example).
on the other hand, bittorrent/ed2k/btsync/syncthing/ipfs provides greate big
file support (and already have many users), but only expose as 
'downloadable files only', lacks a website like user friendly frontend. why 
not join these two greate parts together? 

ipfs client has a http front, makes any ipfs resource accessible through normal
http protocol, makes ipfs resource easily embedded into any existing webpage
without hard work. on the other hand, browser provides native read ability 
for most common file type, makes p2p streaming possible.

inspired by this, here comes my universal p2p project: expose a 
universal http front for multi p2p backend , makes all these p2p resource 
http streamable for any p2p website.

I did not invent any thing new ,just makes better usage for existing things.

```text



   can be any                   agent executable             can be any p2p
http exposed website            runs locally               file share backend
+------------------+        +--------------------+        +-----------------+ 
|   ZeroNet/i2p    |        |  universal p2p     |        | bittorrent/ipfs | 
|  freenet/twister +-------->  http stream front +-------->   triblr/ed2k   | 
|                  |        |                    |        | btsync/syncthing| 
+------------------+        +--------------------+        +-----------------+ 
     
      better support big file access by exposing a universal http front

```


for example, when our agent is running on 127.0.0.1:7788 , you can embedded
a video which stream video from following url:

http://127.0.0.1:7788/bittorrent/v0/stream/c12fe1c06bba254a9dc9f519b335aa7c1367a88a/video.mp4


or simply without specified filename ,just stream the biggest file(most case for video torrent):

http://127.0.0.1:7788/bittorrent/v0/stream/c12fe1c06bba254a9dc9f519b335aa7c1367a88a
(this ideal comes from go-peerflix)

or even embedd a image from archiver as:

http://127.0.0.1:7788/bittorrent/v0/stream/c12fe1c06bba254a9dc9f519b335aa7c1367a88a/a.zip/1.jpg

these file will be download,unpacked by our agent, and stream as these urls(not implemented currently).


with different backend ,we can also access any p2p protocol file ,for example ipfs backend(not implemented currently):

http://127.0.0.1:7788/ipfs/v0/stream/QmarHSr9aSNaPSR6G9KFPbuLV9aEqJfTk1y9B8pdwqK4Rq/myfile.mp3



to archive this goal , our universal http front should fulfill following restrictions:

0. runs headless, quietly, makes user not even mention here is a stream client
1. easily embedded, access resource through http .(this is what we want to do)
2. embed original p2p download link into access url, avoid any user interactive control.
(so user access resource from webpage seamlessly, not mention this is a special p2p resource)
3. every p2p protocol should have access prefix, so they will not clash each other.
4. besides top p2p protocol prefix, every protocol should also have a sub prefix.
  to support protocol upgrade easily (ie, bittorrent can use torrent file or 
  magnet link,btsync has 1.4/2.0 protocol)
5. can access sub-resource. ie, access individue file in torrent . or access a
  file from archives (many resource packed as zip/rar file)
6. user manageable,should expose a web front to view basic disk usage.
7. save disk usage, should have "access from" trace and auto recycle


not goals:(at least at beginning)

1. be a full feature p2p download control client
                

considerations:


why not use magnet link?
magnet provides a universal URI for different p2p protocol already, but not 
all p2p use it, so I'd prefer a simple way.

security?
one site should only view/control p2p content of itself, but universal_p2p can not
knows which site current stream request from . I think any site should ask
universal_p2p for a token (for example, access http://127.0.0.1:8888/get_token)
and access/control p2p content with this token.but this is not straightforward 
for simple usage. so currently only expose "stream" , all GC controlled by 
universal_p2p. 

another problem is that universal_p2p should track all site usage ,but request
is made from client browser,not site itself. so there is no way to tell which 
resource the site belongs to.this problem left as open. 



currently bittorrent streaming function is implemented, most logic comes from go-peerflix

