package segvdl

import (
	"fmt"
	"github.com/pkg/errors"
	"io"
	"log"
	"os/exec"
	"regexp"
)

var SegmentRegexp = regexp.MustCompile(`^(.*segment-)(\d+)(.*)$`)

type Pattern struct {
	Prefix string
	Ext string
}

func (p *Pattern) Parse(s string) error {
	m := SegmentRegexp.FindStringSubmatch(s)
	if len(m) <= SegmentRegexp.NumSubexp() {
		return fmt.Errorf("no regexp match for %s", s)
	}
	p.Prefix = m[1]
	p.Ext = m[3]
	return nil
}

func (p *Pattern) Get(i int) string {
	return fmt.Sprintf("%s%d%s", p.Prefix, i, p.Ext)
}

type Segment struct {
	Order int
	Data io.Reader
}

func Mux(out, video, audio string) error {
	// Video needs putting in non-streaming container
	cmd := exec.Command("ffmpeg", []string{
		"-y", // force overwrite
		"-i", video,
		"-i", audio,
		"-c", "copy", out,
	}...)
	log.Printf("[mux] %s", cmd.String())
	if err := cmd.Run(); err != nil {
		log.Println(cmd.CombinedOutput())
		return errors.Wrap(err, "ffmpeg error")
	}
	return nil
}
