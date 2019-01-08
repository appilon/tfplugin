skip_provider () {
    # declare array of providers to skip
    declare -a arr=("skip1" "skip2")
    for i in "${arr[@]}"
    do
        if [ "terraform-provider-$i" = "$1" ]
        then
            echo "skip $1"
            return
        fi
    done
}