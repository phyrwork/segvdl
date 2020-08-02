# segvdl

## Dependencies

```
# User path
ffmpeg

# Go path
golang.org/x/sync/errgroup
```

## Usage

Play the video you want to download and use your browser's network monitor to identify one of each of the audio and video segments (the video must be playing while the monitor is open).

Filter by *.m4s to quickly find the audio and video segments.

e.g.
https://131vod-adaptive.akamaized.net/exp=1596380909~acl=%2Fb80d81e3-5cfd-4de6-89b6-77d5b7c2570a%2F%2A~hmac=c06f35027eaa1ea5949cb810245d036a9d90c70b7bfe3662f36bdd37f1183984/b80d81e3-5cfd-4de6-89b6-77d5b7c2570a/sep/video/61b70287/chop/segment-72.m4s
https://131vod-adaptive.akamaized.net/exp=1596380909~acl=%2Fb80d81e3-5cfd-4de6-89b6-77d5b7c2570a%2F%2A~hmac=c06f35027eaa1ea5949cb810245d036a9d90c70b7bfe3662f36bdd37f1183984/b80d81e3-5cfd-4de6-89b6-77d5b7c2570a/sep/audio/61b70287/chop/segment-72.m4s

(Note: Sometimes more than just the 'audio' and 'video' subpaths are changed)

```
cd cmd/segvdl
go install
segvdl <any_video_segment> <any_audio_segment> <out_file>
```