NS="cbcontainers-edr-sensor-cleaners"

kubectl get namespace $NS > /dev/null 2>&1
if [ $? -eq 0 ]; then
    kubectl delete namespace $NS
fi

echo "###### Creating $NS namespace ######"
kubectl create namespace $NS

COUNTER=0
for node in $(kubectl get nodes --no-headers -o custom-columns=":metadata.name");
do
    echo "###### Creating cbcontainers-edr-sensor-cleaner-$COUNTER ######"
    echo "apiVersion: batch/v1
kind: Job
metadata:
  name: cbcontainers-edr-sensor-cleaner-$COUNTER
  namespace: $NS
spec:
  template:
    spec:
      volumes:
      - hostPath:
          path: /var/opt
          type: Directory
        name: opt-dir
      containers:
      - name: edr-sensor-cleaner
        image: photon:4.0
        imagePullPolicy: IfNotPresent
        securityContext:
          privileged: true
          runAsUser: 0
        volumeMounts:
        - mountPath: /var/opt
          name: opt-dir
        command: [/bin/sh, -c]
        args:
        - rm -rf /var/opt/carbonblack
      restartPolicy: OnFailure
      nodeName: "$node | kubectl apply -f -
      COUNTER=$(expr $COUNTER + 1)
done

COUNTER=$(expr "$COUNTER" - 1)

for (( c="$COUNTER"; c>=0; c-- ))
do 
   echo "###### Wait for job/cbcontainers-edr-sensor-cleaner-$c to finish ######"
   kubectl -n $NS wait --for=condition=complete --timeout=60s job/cbcontainers-edr-sensor-cleaner-"$c"
done

echo "###### Deleting $NS namespace ######"
kubectl delete namespace $NS

echo "###### DONE ######"
