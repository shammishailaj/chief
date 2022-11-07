package couchbase

import (
	"github.com/couchbase/gocb"
	cbconfig "github.com/shammishailaj/chief/nosql/couchbase/config"
	"os"
	"strconv"
)

type Couchbase struct {
	Conn    *gocb.Cluster
	ConnErr error
	Conf    *cbconfig.Couchbaseconfig
}

func (c *Couchbase) Connect() error {
	c.Conf = new(cbconfig.Couchbaseconfig)
	c.Conf.UserName = os.Getenv("CB_USERNAME")
	c.Conf.Password = os.Getenv("CB_PASSWORD")
	c.Conf.Port, _ = strconv.ParseInt(os.Getenv("CB_PORT"), 10, 64)
	c.Conf.HostName = os.Getenv("CB_HOSTNAME")
	c.Conf.Bucket = os.Getenv("CB_BUCKET")
	c.Conf.BucketPass = os.Getenv("CB_BUCKET_PASS")
	c.Conn, c.ConnErr = c.Conf.Connect()
	return c.ConnErr

}

func (c *Couchbase) OpenBucket() (*gocb.Bucket, error) {
	// Open Bucket
	bucket, err := c.Conn.OpenBucket(c.Conf.Bucket, c.Conf.BucketPass)
	return bucket, err
}

func (c *Couchbase) Upsert(key string, value interface{}) (gocb.Cas, error) {
	bucket, bucketErr := c.OpenBucket()
	if bucketErr != nil {
		return 0, bucketErr
	}
	cas, casErr := bucket.Upsert(key, value, 0)
	return cas, casErr
}
