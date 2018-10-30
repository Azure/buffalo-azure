module github.com/Azure/buffalo-azure

require (
	github.com/Azure/azure-sdk-for-go v18.0.0+incompatible
	github.com/Azure/buffalo-azure/sdk v0.1.0
	github.com/Azure/go-autorest v10.12.0+incompatible
	github.com/gobuffalo/buffalo v0.13.0
	github.com/gobuffalo/buffalo-plugins v1.0.4
	github.com/gobuffalo/makr v1.1.5
	github.com/gobuffalo/pop v4.8.4+incompatible
	github.com/joho/godotenv v1.3.0
	github.com/markbates/inflect v1.0.1
	github.com/marstr/collection v0.3.3 // indirect
	github.com/marstr/randname v0.0.0-20180611202505-48a63b6052f1
	github.com/mitchellh/go-homedir v1.0.0
	github.com/sirupsen/logrus v1.1.1
	github.com/spf13/cobra v0.0.3
	github.com/spf13/viper v1.2.1
)

replace github.com/Azure/buffalo-azure/sdk => ./sdk
