package importer

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "unsafe" // required by //go:linkname
)

// biliHTTPClient aliases the unexported package-level http client used by
// bilibili.Client.doGet so this test can inject a deterministic mock transport.
//
//go:linkname biliHTTPClient mybilibili/bili-proxy/internal/bilibili.httpClient
var biliHTTPClient *http.Client

type importerRoundTripFunc func(req *http.Request) (*http.Response, error)

func (f importerRoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) { return f(req) }

func installImporterBiliClient(t *testing.T, handler importerRoundTripFunc) {
	t.Helper()
	orig := biliHTTPClient
	biliHTTPClient = &http.Client{Transport: handler}
	t.Cleanup(func() { biliHTTPClient = orig })
}

func importerJSONResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

// TestRun_CountsNoopImportAsImported pins the miscount bug in Importer.Run:
// Run increments `imported` whenever importOne returns nil, but importOne
// returns nil WITHOUT importing when the view carries AID==0 or CID==0
// (importer.go:172-174). So a no-op video is reported as "imported" although
// no manuscript/video row was inserted.
func TestRun_CountsNoopImportAsImported(t *testing.T) {
	installImporterBiliClient(t, func(req *http.Request) (*http.Response, error) {
		// View with AID=0 and CID=0 -> importOne is a silent no-op.
		return importerJSONResponse(http.StatusOK,
			`{"code":0,"data":{"bvid":"BV1xx0000000","aid":0,"cid":0,"title":"Noop"}}`), nil
	})

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	defer db.Close()

	// bvExists: video not present -> proceed to GetView.
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM manuscripts WHERE bvid`).WillReturnRows(
		sqlmock.NewRows([]string{"count"}).AddRow(0),
	)

	imp := New(db, Config{})
	imported, skipped, err := imp.Run([]string{"BV1xx0000000"})
	require.NoError(t, err)

	// FINDING-033 修复：无 AID/CID 计为 skipped（不再误计 imported）。
	assert.Equal(t, 0, imported, "FINDING-033 修复: 无 AID/CID 不应计为 imported")
	assert.Equal(t, 1, skipped)

	// Proof importOne never touched the DB: only bvExists's COUNT query ran.
	require.NoError(t, mock.ExpectationsWereMet())
}
