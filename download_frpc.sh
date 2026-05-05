VERSION=0.68.1
wget https://github.com/fatedier/frp/releases/download/v${VERSION}/frp_${VERSION}_linux_amd64.tar.gz
tar -zxvf frp_${VERSION}_linux_amd64.tar.gz
mv frp_${VERSION}_linux_amd64 frp
rm frp_${VERSION}_linux_amd64.tar.gz
