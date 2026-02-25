package main

import (
	"bytes"
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	endpoint := getEnv("MINIO_ENDPOINT", "minio:9000")
	accessKey := getEnv("MINIO_ACCESS_KEY", "minioadmin")
	secretKey := getEnv("MINIO_SECRET_KEY", "minioadmin")
	bucketName := getEnv("BUCKET_NAME", "snapshots")
	objectName := getEnv("OBJECT_NAME", "doc1-v1.json")
	mitigated, _ := strconv.ParseBool(getEnv("MITIGATED", "false"))

	ctx := context.Background()

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalf("Neuspesno povezivanje na MinIO: %v", err)
	}

	exists, err := client.BucketExists(ctx, bucketName)
	if err != nil {
		log.Fatal(err)
	}
	if !exists {
		err = client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			log.Fatalf("Neuspesno kreiranje bucket-a: %v", err)
		}
		log.Printf("Kreiran bucket: %s", bucketName)
	}

	if mitigated {
		err = client.EnableVersioning(ctx, bucketName)
		if err != nil {
			log.Fatalf("Neuspesno ukljucivanje versioninga: %v", err)
		}
		log.Println("MITIGOVANA VERZIJA: Versioning ukljucen na bucket-u")
	} else {
		log.Println("RANJIVA VERZIJA: Bucket bez versioninga")
	}

	originalContent := `{"version":1,"data":"original","timestamp":"` + time.Now().String() + `"}`
	info, err := client.PutObject(ctx, bucketName, objectName,
		bytes.NewReader([]byte(originalContent)), int64(len(originalContent)),
		minio.PutObjectOptions{ContentType: "application/json"})
	if err != nil {
		log.Fatalf("Greska pri upload-u originala: %v", err)
	}
	log.Printf("Original snapshot uploadovan. ETag: %s, VersionID: %s", info.ETag, info.VersionID)

	time.Sleep(1 * time.Second)

	maliciousContent := `{"version":1,"data":"kompromitovano","timestamp":"` + time.Now().String() + `"}`
	info2, err := client.PutObject(ctx, bucketName, objectName,
		bytes.NewReader([]byte(maliciousContent)), int64(len(maliciousContent)),
		minio.PutObjectOptions{ContentType: "application/json"})
	if err != nil {
		log.Fatalf("Greska pri upload-u malicioznog sadrzaja: %v", err)
	}
	log.Printf("Maliciozni snapshot uploadovan. ETag: %s, VersionID: %s", info2.ETag, info2.VersionID)

	obj, err := client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		log.Fatal(err)
	}
	defer obj.Close()

	buf := new(bytes.Buffer)
	buf.ReadFrom(obj)
	log.Printf("Trenutni sadrzaj objekta: %s", buf.String())

	log.Println("Verzije objekta:")
	opts := minio.ListObjectsOptions{
		Prefix:       objectName,
		WithVersions: true,
	}
	versionCount := 0
	for objInfo := range client.ListObjects(ctx, bucketName, opts) {
		if objInfo.Err != nil {
			log.Fatal(objInfo.Err)
		}
		log.Printf("  VersionID: %s, IsLatest: %v, LastModified: %v", objInfo.VersionID, objInfo.IsLatest, objInfo.LastModified)
		versionCount++
	}
	if versionCount == 0 {
		log.Println("  (nema versioninga ili nema verzija)")
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
