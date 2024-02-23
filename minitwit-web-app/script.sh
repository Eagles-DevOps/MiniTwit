echo "-----------------------------------"
echo "         Script start"
echo "-----------------------------------"

# Download Go tarball
wget -q https://golang.org/dl/go1.22.0.linux-amd64.tar.gz
# Extract tarball
sudo tar -C /usr/local -xzf go1.22.0.linux-amd64.tar.gz
# Add Go binary directory to PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> $HOME/.profile
export PATH=$PATH:/usr/local/go/bin

echo "-----------------------------------"
echo "         Golang installed"
echo "-----------------------------------"

# Set GOPATH; adjust this as needed
echo "export GOPATH=$HOME/go" >> $HOME/.profile
source $HOME/.profile

# Copy or clone your application files to the VM
git clone https://github.com/Eagles-DevOps/MiniTwit.git

echo "-----------------------------------"
echo "          Repo clonned"
echo "-----------------------------------"

# Navigate to your app directory and build/run your Go app
cd MiniTwit/minitwit-api


echo "-----------------------------------"
echo "   Downloading / Running app .... "
echo "-----------------------------------"
go run minitwit-api.go