skip_provider () {
    # declare array of providers to skip
    declare -a arr=("aws" "oci" "azurerm" "google" "google-beta" "kubernetes")
    for i in "${arr[@]}"
    do
        if [ "terraform-provider-$i" = "$1" ]
        then
            echo "skip $1"
            return
        fi
    done
}