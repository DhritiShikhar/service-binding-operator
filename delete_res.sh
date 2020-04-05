oc delete sa service-binding-operator;
oc delete crd servicebindingrequests.apps.openshift.io;
oc delete role service-binding-operator;
oc delete rolebinding service-binding-operator;
oc delete secret scorecard-kubeconfig;
operator-sdk scorecard
