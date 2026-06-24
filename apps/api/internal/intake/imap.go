package intake

import (
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

// IMAPConfig configures the IMAP message source.
type IMAPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Mailbox  string // mailbox to scan, e.g. "INBOX"
	From     string // sender filter (IMAP SEARCH HEADER FROM), e.g. "99freelas.com.br"
	Limit    uint32 // newest N messages per poll; 0 = default
}

const (
	imapDefaultLimit   = 50
	imapCommandTimeout = 30 * time.Second
)

// IMAPSource fetches raw messages from an IMAP mailbox over implicit TLS. Each
// Fetch opens a fresh read-only session (EXAMINE — \Seen is never set), searches
// by sender, downloads the matches with BODY.PEEK, and logs out. Read-only by
// construction: it never deletes or flags anything in the mailbox.
type IMAPSource struct{ cfg IMAPConfig }

var _ MessageSource = (*IMAPSource)(nil)

// NewIMAPSource builds an IMAP-backed message source.
func NewIMAPSource(cfg IMAPConfig) *IMAPSource {
	if cfg.Limit == 0 {
		cfg.Limit = imapDefaultLimit
	}
	if cfg.Mailbox == "" {
		cfg.Mailbox = "INBOX"
	}
	return &IMAPSource{cfg: cfg}
}

// Fetch returns the raw RFC 822 bytes of the most recent messages from the
// configured sender.
func (s *IMAPSource) Fetch(ctx context.Context) ([][]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	addr := net.JoinHostPort(s.cfg.Host, strconv.Itoa(s.cfg.Port))
	c, err := client.DialTLS(addr, nil)
	if err != nil {
		return nil, fmt.Errorf("intake imap: dial %s: %w", addr, err)
	}
	c.Timeout = imapCommandTimeout
	defer func() { _ = c.Logout() }()

	if err := c.Login(s.cfg.Username, s.cfg.Password); err != nil {
		return nil, fmt.Errorf("intake imap: login: %w", err)
	}
	if _, err := c.Select(s.cfg.Mailbox, true); err != nil { // read-only (EXAMINE)
		return nil, fmt.Errorf("intake imap: select %q: %w", s.cfg.Mailbox, err)
	}

	criteria := imap.NewSearchCriteria()
	if s.cfg.From != "" {
		criteria.Header.Add("From", s.cfg.From)
	}
	nums, err := c.Search(criteria)
	if err != nil {
		return nil, fmt.Errorf("intake imap: search: %w", err)
	}
	if len(nums) == 0 {
		return nil, nil
	}
	if uint32(len(nums)) > s.cfg.Limit { // keep the newest N (search returns ascending)
		nums = nums[uint32(len(nums))-s.cfg.Limit:]
	}

	seqset := new(imap.SeqSet)
	seqset.AddNum(nums...)
	section := &imap.BodySectionName{Peek: true} // BODY.PEEK[] — does not set \Seen

	messages := make(chan *imap.Message, len(nums))
	fetchErr := make(chan error, 1)
	go func() {
		fetchErr <- c.Fetch(seqset, []imap.FetchItem{section.FetchItem()}, messages)
	}()

	var out [][]byte
	for msg := range messages {
		body := msg.GetBody(section)
		if body == nil {
			continue
		}
		b, err := io.ReadAll(body)
		if err != nil {
			return nil, fmt.Errorf("intake imap: read body: %w", err)
		}
		out = append(out, b)
	}
	if err := <-fetchErr; err != nil {
		return nil, fmt.Errorf("intake imap: fetch: %w", err)
	}
	return out, nil
}
