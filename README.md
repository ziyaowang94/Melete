There are great differences in consensus process, block structure, execution order, and broadcast mode, which makes it difficult to reuse modules and introduces many redundant designs
So one folder per  system for now

message format:
channel_id
message
messageType

Every time you add a new message, be sure to register the message type in the definition folder
The message type is directly determined by messageType and deserialized by message.

Proto Installation Tutorial: proto/README.md

This project also requires rocksdb v6.28.2 to be installed 

How to run Uranus and experiment can be found in start.md