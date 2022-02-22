for ((n=0;n<1000;n++))
do
KUBECONFIG=/tmp/kc-test.yaml kubectl get all -A
done