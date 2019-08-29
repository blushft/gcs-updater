package main

import (
	"context"
	"flag"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/davecgh/go-spew/spew"

	"google.golang.org/api/iterator"

	"cloud.google.com/go/storage"
	"github.com/fsouza/fake-gcs-server/fakestorage"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"google.golang.org/api/option"
)

var conf = config{
	port:     "5050",
	bucket:   "dist",
	credfile: "seviceacct.json",
}

var (
	fakerF = flag.Bool("fake", true, "enable fake gcs server")
)

type config struct {
	port     string
	bucket   string
	credfile string
}

func main() {
	ctx := context.Background()
	var client *storage.Client
	var server *fakestorage.Server
	var err error
	if *fakerF {
		server, err = fakestorage.NewServerWithOptions(fakestorage.Options{
			InitialObjects: []fakestorage.Object{
				{
					BucketName: conf.bucket,
					Name:       "just-something",
					Content:    []byte("I HAS A BUCKET"),
				},
			},
			StorageRoot: "./testapp",
		})
		if err != nil {
			log.Fatal(err)
		}
		objs, keys, err := server.ListObjects(conf.bucket, "*", "")
		if err != nil {
			log.Fatal(err)
		}
		spew.Dump(objs, keys)
		client = server.Client()
	} else {
		client, err = storage.NewClient(ctx, option.WithCredentialsFile(conf.credfile))
		if err != nil {
			log.Fatal(err)
		}
	}

	bucket := client.Bucket(conf.bucket)
	log.Println(bucket)

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/list", func(c echo.Context) error {
		ret := make(map[string]string)
		it := bucket.Objects(ctx, nil)
		for {
			attrs, err := it.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				c.Logger().Error(err)
				return c.NoContent(http.StatusInternalServerError)
			}
			ret[attrs.Name] = attrs.ContentType
		}
		return c.JSON(http.StatusOK, ret)
	})

	e.GET("/*", func(c echo.Context) error {
		start := time.Now()
		obj := bucket.Object(c.Request().URL.Path[1:])
		attr, err := obj.Attrs(ctx)
		if err != nil {
			c.Logger().Error(err)
			return c.NoContent(http.StatusNotFound)
		}

		artifact := obj.ReadCompressed(true)
		r, err := artifact.NewReader(ctx)
		if err != nil {
			c.Logger().Error(err)
			return c.NoContent(http.StatusNotFound)
		}
		defer r.Close()

		c.Response().Header().Set("Content-Type", attr.ContentType)
		c.Response().Header().Set("Content-Encoding", attr.ContentEncoding)
		c.Response().Header().Set("Content-Length", strconv.Itoa(int(attr.Size)))
		c.Response().WriteHeader(http.StatusOK)
		if _, err := io.Copy(c.Response().Writer, r); err != nil {
			return err
		}
		c.Response().Flush()

		c.Logger().Infof("DL %s in %d", obj.ObjectName, time.Since(start))
		return nil
	})

	e.Logger.Fatal(e.Start(":" + conf.port))
}
