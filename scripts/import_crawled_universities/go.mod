module github.com/gaokaohub/gaokao/scripts/import_crawled_universities

go 1.23.8

replace github.com/gaokaohub/gaokao/pkg/scripts => ../../pkg/scripts

require (
	github.com/gaokaohub/gaokao/pkg/scripts v0.0.0-00010101000000-000000000000
	github.com/lib/pq v1.10.9
)
