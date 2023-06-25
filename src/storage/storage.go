package storage

// this is a snippet from MongoDB, will need work

/* func connectMongo() {
	ctx := context.TODO()
	uri := "mongodb+srv://dtools-mos01.ev9rirj.mongodb.net/?authSource=%24external&authMechanism=MONGODB-X509&retryWrites=true&w=majority&tlsCertificateKeyFile=<path_to_certificate>"
	serverAPIOptions := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().
		ApplyURI(uri).
		SetServerAPIOptions(serverAPIOptions)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Disconnect(ctx)

	collection := client.Database("mos01-dev").Collection("mosobj01")

	docCount, err := collection.CountDocuments(ctx, bson.D{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(docCount)
}
*/
