package provider

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"movie-data-api/server"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"gorm.io/gorm"
)

type S3ItemMapping struct {
	Name         string    `json:"name"`
	LastModified time.Time `json:"last_modified"`
	Size         int64     `json:"size"`
}

type AWSProvider struct {
	IProvider
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	Region          string
	Interval        time.Duration
	S3Service       *s3.S3
	db              *gorm.DB
}

func NewAWSProvider(accessKeyID string, secretAccessKey string, bucketName string, region string,
	db *gorm.DB) (*AWSProvider, error) {

	sess, err := session.NewSession()

	return &AWSProvider{
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		BucketName:      bucketName,
		Region:          region,
		Interval:        time.Second * 15,
		S3Service:       s3.New(sess),
		db:              db,
	}, err
}

func (p *AWSProvider) Run() error {

	for {
		resp, err := p.S3Service.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(p.BucketName)})
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			continue
		}

		for _, item := range resp.Contents {
			s3ItemMapping := S3ItemMapping{
				Name:         *item.Key,
				LastModified: *item.LastModified,
				Size:         *item.Size,
			}
			fmt.Println("Name:         ", *item.Key)
			fmt.Println("Last modified:", *item.LastModified)
			fmt.Println("Size:         ", *item.Size)
			fmt.Println("Storage class:", *item.StorageClass)
			fmt.Println("")

			var foundMapping S3ItemMapping
			tx := p.db.Find(&server.Movie{Title: *item.Key}).First(&foundMapping)

			// if foundMapping.Name != "" {
			// 	tx = p.db.Where("name = ?", foundMapping.Name).Delete(&s3ItemMapping)
			// 	if tx.Error != nil {
			// 		fmt.Printf("Error: %s\n", tx.Error)
			// 		continue
			// 	}
			// }

			if tx.Error != nil && tx.Error == gorm.ErrRecordNotFound || foundMapping.Name == "" {

				// Index the item -----------------------------------------------

				// Read the S3 item into memory
				result, err := p.S3Service.GetObject(&s3.GetObjectInput{
					Bucket: aws.String(p.BucketName),
					Key:    aws.String(*item.Key),
				})
				if err != nil {
					fmt.Println(err)
					continue
				}
				b, err := ioutil.ReadAll(result.Body)
				if err != nil {
					fmt.Println(err)
					continue
				}
				var movie server.Movie
				// defer wont work in a loop
				result.Body.Close()
				err = json.Unmarshal(b, &movie)
				if err != nil {
					fmt.Println(err)
					continue
				}
				var foundMovie server.Movie
				p.db.Where("title = ?", movie.Title).Find(&foundMovie)
				if foundMovie.Title == "" {
					// If we don't have the movie in the database, add it
					tx = p.db.Create(&movie)
					if tx.Error != nil {
						fmt.Println(tx.Error)
						continue
					}
				}
				// Index the itemMapping -------------------------------------
				tx := p.db.Create(&s3ItemMapping)
				if tx.Error != nil {
					fmt.Printf("Error: %s\n", tx.Error)
					continue
				}

			}

		}
		time.Sleep(p.Interval)
	}

}
