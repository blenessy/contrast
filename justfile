# Undeploy, rebuild, deploy.
default target=default_deploy_target: undeploy coordinator initializer (deploy target)

# Build the coordinator, containerize and push it.
coordinator:
    nix run .#push-coordinator -- "$container_registry/coordinator:latest"

# Build the initializer, containerize and push it.
initializer:
    nix run .#push-initializer -- "$container_registry/initializer:latest"

default_deploy_target := "simple"
worspace_dir := "workspace"

# Generate policies, apply Kubernetes manifests.
deploy target=default_deploy_target: (generate target) apply

# Generate policies, update manifest.
generate target=default_deploy_target:
    #!/usr/bin/env bash
    mkdir -p ./{{worspace_dir}}
    rm -rf ./{{worspace_dir}}/deployment
    cp -R ./deployments/{{target}} ./{{worspace_dir}}/deployment
    cp ./data/manifest.json ./{{worspace_dir}}/manifest.json
    nix run .#yq-go -- -i ". \
        | with(select(.spec.template.spec.containers[].image | contains(\"coordinator\")); \
        .spec.template.spec.containers[0].image = \"${container_registry}/coordinator:latest\")" \
        ./{{worspace_dir}}/deployment/coordinator.yml
    for f in ./{{worspace_dir}}/deployment/*.yml; do
        nix run .#yq-go -- -i ". \
            | with(select(.spec.template.spec.initContainers[].image | contains(\"initializer\")); \
            .spec.template.spec.initContainers[0].image = \"${container_registry}/initializer:latest\")" \
            "${f}"
    done
    nix run .#cli -- generate \
        -m ./{{worspace_dir}}/manifest.json \
        -p tools \
        -s genpolicy-msft.json \
        ./{{worspace_dir}}/deployment/*.yml

# Apply Kubernetes manifests from /deployment
apply:
    kubectl apply -f ./{{worspace_dir}}/deployment/ns.yml
    kubectl apply -f ./{{worspace_dir}}/deployment

# Delete Kubernetes manifests.
undeploy:
    -kubectl delete -f ./{{worspace_dir}}/deployment

# Create a CoCo-enabled AKS cluster.
create:
    nix run .#create-coco-aks -- --name="$azure_resource_group"

# Set the manifest at the coordinator.
set:
    #!/usr/bin/env bash
    kubectl port-forward pod/port-forwarder-coordinator 1313 &
    PID=$!
    sleep 1
    nix run .#cli -- set \
        -m ./{{worspace_dir}}/manifest.json \
        -c localhost:1313 \
        ./{{worspace_dir}}/deployment/*.yml
    kill $PID

# Load the kubeconfig from the running AKS cluster.
get-credentials:
    nix run .#azure-cli -- aks get-credentials \
        --resource-group "$azure_resource_group" \
        --name "$azure_resource_group"

# Destroy a running AKS cluster.
destroy:
    nix run .#destroy-coco-aks -- "$azure_resource_group"

# Run code generators.
codegen:
    nix run .#generate

# Format code.
fmt:
    nix fmt

# Cleanup auxiliary files, caches etc.
clean:
    rm -rf ./{{worspace_dir}}

# Template for the justfile.env file.
rctemplate := '''
# Container registry to push images to
container_registry=""
# Azure resource group/ resource name. Resource group will be created.
azure_resource_group=""
'''

# Developer onboarding.
onboard:
    @ [[ -f "./justfile.env" ]] && echo "justfile.env already exists" && exit 1 || true
    @echo '{{ rctemplate }}' > ./justfile.env
    @echo "Created ./justfile.env. Please fill it out."

# Just configuration.
set dotenv-filename := "justfile.env"
set dotenv-load := true
set shell := ["bash", "-uc"]