#!/bin/bash

set -e
usage() {
  echo "Usage: $0 [flags]"
  echo "OPTIONS"
  echo -e "\t-u, --user <user:password>"
  echo -e "\t\tSpecify the user name and password to use for server authentication"
  echo -e "\t\tIf you simply specify the user name, curl will prompt for a password"
  echo -e "\t--no-input"
  echo -e "\t\tExplicitly disable interactivity when running in a terminal"
  echo -e "\t-h, --help"
  echo -e "\t\thelp for install-kconnect.sh"
}
install_kconnect() {
  cd ..
  INSTALL_PATH=$1
  if [[ $NO_INPUT != "YES" ]]; then
    read -p "Install in default path, $INSTALL_PATH? (Y/N): " confirm
    if [[ $confirm != [yY] && $confirm != [yY][eE][sS] ]]; then
      read -p "Please provide install path: " path
      INSTALL_PATH=$path
    fi
  fi
  if [[ ! -d $INSTALL_PATH ]]; then
    mkdir -p $INSTALL_PATH
  fi
  if [[ -d $INSTALL_PATH/$KCONNECT_DIR ]]; then
    rm -rf $INSTALL_PATH/$KCONNECT_DIR
  fi
  mv $KCONNECT_DIR $INSTALL_PATH
}

USER=""
NO_INPUT="NO"
while [[ $# -gt 0 ]]; do
  case $1 in
  -u | --user)
    USER="$1 $2"
    shift # past argument
    shift # past value
    ;;
  -h | --help)
    usage
    exit 0
    ;;
  --no-input)
    NO_INPUT="YES"
    shift # past argument
    ;;
  -* | --*)
    echo "ERROR: Unknown option $1"
    exit 1
    ;;
  esac
done

echo "creating directory kconnect"
KCONNECT_DIR="kconnect"
mkdir -p $KCONNECT_DIR
cd $KCONNECT_DIR

latest_kconnect_release_tag=$(curl -fsSLI -o /dev/null -w %{url_effective} $USER https://github.com/fidelity/kconnect/releases/latest | sed 's#.*/##')
latest_kubectl_release_tag=$(curl -k -L --silent $USER https://dl.k8s.io/release/stable.txt)
latest_helm_release_tag=$(curl -fsSLI -o /dev/null -w %{url_effective} $USER https://github.com/helm/helm/releases/latest | sed 's#.*/##')
latest_kubelogin_release_tag=$(curl -fsSLI -o /dev/null -w %{url_effective} $USER https://github.com/Azure/kubelogin/releases/latest | sed 's#.*/##')
latest_aws_iam_authenticator_release_tag=$(curl -fsSLI -o /dev/null -w %{url_effective} $USER https://github.com/kubernetes-sigs/aws-iam-authenticator/releases/latest | sed 's#.*/##' | cut -c2-)
latest_azure_cli_release_tag=$(curl -fsSLI -o /dev/null -w %{url_effective} $USER https://github.com/Azure/azure-cli/releases/latest | sed 's#.*/##')

echo "kconnect version: $latest_kconnect_release_tag"
echo "kubectl version: $latest_kubectl_release_tag"
echo "helm version: $latest_helm_release_tag"
echo "kubelogin version: $latest_kubelogin_release_tag"
echo "aws-iam-authenticator version: $latest_aws_iam_authenticator_release_tag"
echo "azure-cli version: $latest_azure_cli_release_tag"

if [[ "$OSTYPE" == "linux-gnu"* ]]; then
  # linux
  #arch=$(dpkg --print-architecture)
  arch_output=$(uname -m)
  arch=""
  case $arch_output in

  x86_64)
    arch="amd64"
    ;;

  aarch64)
    arch="arm64"
    ;;

  aarch)
    arch="arm"
    ;;

  esac

  echo "arch: " $arch

  kconnect_url=$(echo "https://github.com/fidelity/kconnect/releases/download/TAG/kconnect_linux_ARCH.tar.gz" | sed "s/TAG/$latest_kconnect_release_tag/" | sed "s/ARCH/$arch/")
  kubectl_url=$(echo "https://dl.k8s.io/release/TAG/bin/linux/ARCH/kubectl" | sed "s/TAG/$latest_kubectl_release_tag/" | sed "s/ARCH/$arch/")
  helm_url=$(echo "https://get.helm.sh/helm-TAG-linux-ARCH.tar.gz" | sed "s/TAG/$latest_helm_release_tag/" | sed "s/ARCH/$arch/")
  aws_iam_authenticator_url=$(echo "https://github.com/kubernetes-sigs/aws-iam-authenticator/releases/download/vTAG/aws-iam-authenticator_TAG_linux_ARCH" | sed "s/TAG/$latest_aws_iam_authenticator_release_tag/g" | sed "s/ARCH/$arch/")
  kubelogin_url=$(echo "https://github.com/Azure/kubelogin/releases/download/TAG/kubelogin-linux-amd64.zip" | sed "s/TAG/$latest_kubelogin_release_tag/")
  azure_url="https://aka.ms/InstallAzureCli"

  echo "kconnect url: $kconnect_url"
  echo "kubectl url: $kubectl_url"
  echo "helm url: $helm_url"
  echo "aws_iam_authenticator url: $aws_iam_authenticator_url"
  echo "kubelogin url: $kubelogin_url"
  echo "azure url: $azure_url"

  # download
  curl -s $USER -L $kconnect_url -o kconnect.tar.gz
  curl -s $USER -LO $kubectl_url
  curl -s $USER -L $helm_url -o helm.tar.gz
  curl -s $USER -L $aws_iam_authenticator_url -o aws-iam-authenticator
  curl -s $USER -L $kubelogin_url -o kubelogin.zip
  curl -s $USER -L $azure_url -o azure-cli-install.sh

  # unzip
  tar -xf kconnect.tar.gz
  tar -xf helm.tar.gz
  mv linux-*/helm .
  unzip -qq kubelogin.zip
  mv bin/linux_amd64/kubelogin .

  # cleanup
  rm -f kconnect.tar.gz
  rm -f helm.tar.gz
  rm -rf linux-*
  rm -f kubelogin.zip
  rm -rf bin

  # permissions
  chmod +x kubectl
  chmod +x aws-iam-authenticator
  chmod +x kubelogin
  chmod +x azure-cli-install.sh

  # install
  install_kconnect "/usr/local/bin"

elif [[ "$OSTYPE" == "darwin"* ]]; then
  # Mac OSX
  kconnect_url=$(echo "https://github.com/fidelity/kconnect/releases/download/TAG/kconnect_macos_amd64.tar.gz" | sed "s/TAG/$latest_kconnect_release_tag/")
  kubectl_url=$(echo "https://dl.k8s.io/release/TAG/bin/darwin/amd64/kubectl" | sed "s/TAG/$latest_kubectl_release_tag/")
  helm_url=$(echo "https://get.helm.sh/helm-TAG-darwin-amd64.tar.gz" | sed "s/TAG/$latest_helm_release_tag/")
  aws_iam_authenticator_url=$(echo "https://github.com/kubernetes-sigs/aws-iam-authenticator/releases/download/vTAG/aws-iam-authenticator_TAG_darwin_amd64" | sed "s/TAG/$latest_aws_iam_authenticator_release_tag/g")
  kubelogin_url=$(echo "https://github.com/Azure/kubelogin/releases/download/TAG/kubelogin-darwin-amd64.zip" | sed "s/TAG/$latest_kubelogin_release_tag/")
  azure_url="https://docs.microsoft.com/en-us/cli/azure/install-azure-cli-macos"

  echo "kconnect url: $kconnect_url"
  echo "kubectl url: $kubectl_url"
  echo "helm url: $helm_url"
  echo "aws_iam_authenticator url: $aws_iam_authenticator_url"
  echo "kubelogin url: $kubelogin_url"
  echo "azure url: $azure_url"

  # download
  curl -s $USER -L $kconnect_url -o kconnect.tar.gz
  curl -s $USER -LO $kubectl_url
  curl -s $USER -L $helm_url -o helm.tar.gz
  curl -s $USER -L $aws_iam_authenticator_url -o aws-iam-authenticator
  curl -s $USER -L $kubelogin_url -o kubelogin.zip

  # unzip
  tar -xf kconnect.tar.gz
  tar -xf helm.tar.gz
  mv darwin-*/helm .
  unzip -qq kubelogin.zip
  mv bin/darwin_amd64/kubelogin .

  # cleanup
  rm -f kconnect.tar.gz
  rm -f helm.tar.gz
  rm -rf darwin-*
  rm -f kubelogin.zip
  rm -rf bin

  # permissions
  chmod +x kubectl
  chmod +x aws-iam-authenticator
  chmod +x kubelogin

  # install
  install_kconnect "/usr/local/bin"

elif [[ "$OSTYPE" == "msys" ]]; then
  # Win git bash
  kconnect_url=$(echo "https://github.com/fidelity/kconnect/releases/download/TAG/kconnect_windows_amd64.zip" | sed "s/TAG/$latest_kconnect_release_tag/")
  kubectl_url=$(echo "https://dl.k8s.io/release/TAG/bin/windows/amd64/kubectl.exe" | sed "s/TAG/$latest_kubectl_release_tag/")
  helm_url=$(echo "https://get.helm.sh/helm-TAG-windows-amd64.zip" | sed "s/TAG/$latest_helm_release_tag/")
  aws_iam_authenticator_url=$(echo "https://github.com/kubernetes-sigs/aws-iam-authenticator/releases/download/vTAG/aws-iam-authenticator_TAG_windows_amd64.exe" | sed "s/TAG/$latest_aws_iam_authenticator_release_tag/g")
  kubelogin_url=$(echo "https://github.com/Azure/kubelogin/releases/download/TAG/kubelogin-win-amd64.zip" | sed "s/TAG/$latest_kubelogin_release_tag/")
  azure_url=$(echo "https://github.com/Azure/azure-cli/releases/download/TAG/TAG.msi" | sed "s/TAG/$latest_azure_cli_release_tag/g")

  echo "kconnect url: $kconnect_url"
  echo "kubectl url: $kubectl_url"
  echo "helm url: $helm_url"
  echo "aws_iam_authenticator url: $aws_iam_authenticator_url"
  echo "kubelogin url: $kubelogin_url"
  echo "azure url: $azure_url"

  # download
  curl -k -s $USER -L $kconnect_url -o kconnect.zip
  curl -k -s $USER -LO $kubectl_url
  curl -k -s $USER -L $helm_url -o helm.zip
  curl -k -s $USER -L $aws_iam_authenticator_url -o aws-iam-authenticator.exe
  curl -k -s $USER -L $kubelogin_url -o kubelogin.zip
  curl -k -s $USER -LO $azure_url

  # unzip
  unzip -qq kconnect.zip
  unzip -qq helm.zip
  mv windows-amd64/helm.exe .
  unzip -qq kubelogin.zip
  mv bin/windows_amd64/kubelogin.exe .

  # cleanup
  rm -f kconnect.zip
  rm -f helm.zip
  rm -rf windows-amd64
  rm -f kubelogin.zip
  rm -rf bin

  # install
  install_kconnect "$HOME/bin"
fi
