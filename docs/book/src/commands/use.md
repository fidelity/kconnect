# kconnect use

## Purpose

The purpose of the **use** command is to discover clusters you have access to in a cluster provider (i.e. EKS, AKS) using your identity which is supplied by a specific IdP.

It will query the cluster provider and get a list of clusters you have access to. When you select a cluster to connect to your kubeconfig will be updated with the connection details for the selected cluster.

