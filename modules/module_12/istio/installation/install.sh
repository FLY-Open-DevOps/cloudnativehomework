curl -L https://istio.io/downloadIstio | sh -
cd istio-*
cp bin/istioctl /usr/local/bin
istioctl install --set profile=demo -y