## Deploy Using Docker from your Local Machine

#### Pre-Requisites

For these instructions to be relevant to your Buffalo-Application you must:
- Both Buffalo and Buffalo-Azure should be available in the PATH of your local machine. If you don't have them you can 
  get started here: 
  - Buffalo: https://gobuffalo.io/en/docs/installation
  - Buffalo-Azure: [Installing Buffalo Azure](./README.md#Installation)
- Have `docker` installed on your local machine. You can find a download here: https://www.docker.com/get-docker
  - If you're using Windows or Linux, Virtualization will have to be enabled on your machine. If it is not enabled, 
  you'll have to reboot and configure your system settings in the BIOS.
- Your Buffalo Application should have a "Dockerfile" in the root of the repository.
- A Docker repository exists to host your image. Don't have one? You can create one quickly following one of these 
guides:
  - [Docker Hub Repository Documentaion](https://docs.docker.com/docker-hub/repos/)
  - [Azure Container Registry Documentation](https://docs.microsoft.com/en-us/azure/container-registry/) 
- You have access to an Azure subscription. ([Don't have one? Get started for free, right now.](https://aka.ms/buffalo-free-account)) 

### Step-By-Step

1. Navigate to the root directory of your project:

  ``` bash
  cd {your project root}
  ```

2. Build a Docker image from your project's defintion file:

  ``` bash
  docker build -t [{repository}/]{application}[:tag] .
  ```

  i.e.

  ``` bash
  docker build -t marstr.azurecr.io/musicvotes:latest .
  ```

3. Authenticate with Docker so that you can push the image to your repository:

  To protect your password from being logged, put it in a file named `mypassword.txt` then pipe it into docker:

  ``` bash
  mypassword.txt | docker login {repository} -u {username} --password-stdin 
  ```

  i.e.

  ``` bash
  mypassword.txt | docker login marstr.azurecr.io -u marstr --password-stdin
  ```

4. Push your image to the Docker repository:

  ``` bash
  docker push {repository}/{application}/{application}[:{tag}]
  ```
 
  i.e.
 
  ``` bash
  docker push marstr.azurecr.io/musicvotes:latest
  ```
  
5. Provision the infrastructure in Azure that will host your web application:

  ``` bash
  buffalo azure provision --subscription {subscriptionID} --image {imageID}
  ```
  
  i.e.
  
  ``` bash
  buffalo azure provision --subscription 659641ac-3e41-4ee7-9eb1-26c84469893d --image marstr.azurecr.io/musicvotes:latest
  ```

  Don't know your Azure subscription ID? A few easy steps in the Portal will solve that: 
  [[blog] Getting your Azure Subscription GUID (new portal)](https://blogs.msdn.microsoft.com/mschray/2016/03/18/getting-your-azure-subscription-guid-new-portal/)
 
You're done! After a few minutes, your site will be live on Azure!