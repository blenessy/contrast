# Copyright 2024 Edgeless Systems GmbH
# SPDX-License-Identifier: AGPL-3.0-only

{
  lib,
  ociLayerTar,
  ociImageManifest,
  ociImageLayout,
  contrast-node-installer,
  kata,
  pkgsStatic,
  writers,
}:

let
  node-installer = ociLayerTar {
    files = [
      {
        source = lib.getExe contrast-node-installer;
        destination = "/bin/node-installer";
      }
      {
        source = "${pkgsStatic.util-linux}/bin/nsenter";
        destination = "/bin/nsenter";
      }
    ];
  };

  launch-digest = lib.removeSuffix "\n" (
    builtins.readFile "${kata.runtime-class-files}/launch-digest.hex"
  );
  runtime-handler = lib.removeSuffix "\n" (
    builtins.readFile "${kata.runtime-class-files}/runtime-handler"
  );

  installer-config = ociLayerTar {
    files = [
      {
        source = writers.writeJSON "contrast-node-install.json" {
          files = [
            {
              url = "file:///opt/edgeless/share/kata-containers.img";
              path = "/opt/edgeless/${runtime-handler}/share/kata-containers.img";
            }
            {
              url = "file:///opt/edgeless/share/kata-kernel";
              path = "/opt/edgeless/${runtime-handler}/share/kata-kernel";
            }
            {
              url = "file:///opt/edgeless/snp/bin/qemu-system-x86_64";
              path = "/opt/edgeless/${runtime-handler}/snp/bin/qemu-system-x86_64";
            }
            {
              url = "file:///opt/edgeless/tdx/bin/qemu-system-x86_64";
              path = "/opt/edgeless/${runtime-handler}/tdx/bin/qemu-system-x86_64";
            }
            {
              url = "file:///opt/edgeless/snp/share/OVMF.fd";
              path = "/opt/edgeless/${runtime-handler}/snp/share/OVMF.fd";
            }
            {
              url = "file:///opt/edgeless/tdx/share/OVMF.fd";
              path = "/opt/edgeless/${runtime-handler}/tdx/share/OVMF.fd";
            }
            {
              url = "file:///opt/edgeless/bin/containerd-shim-contrast-cc-v2";
              path = "/opt/edgeless/${runtime-handler}/bin/containerd-shim-contrast-cc-v2";
            }
            {
              url = "file:///opt/edgeless/bin/kata-runtime";
              path = "/opt/edgeless/${runtime-handler}/bin/kata-runtime";
            }
            {
              url = "file:///opt/edgeless/snp/share/qemu/kvmvapic.bin";
              path = "/opt/edgeless/${runtime-handler}/snp/share/qemu/kvmvapic.bin";
            }
            {
              url = "file:///opt/edgeless/snp/share/qemu/linuxboot_dma.bin";
              path = "/opt/edgeless/${runtime-handler}/snp/share/qemu/linuxboot_dma.bin";
            }
            {
              url = "file:///opt/edgeless/snp/share/qemu/efi-virtio.rom";
              path = "/opt/edgeless/${runtime-handler}/snp/share/qemu/efi-virtio.rom";
            }
            {
              url = "file:///opt/edgeless/tdx/share/qemu/kvmvapic.bin";
              path = "/opt/edgeless/${runtime-handler}/tdx/share/qemu/kvmvapic.bin";
            }
            {
              url = "file:///opt/edgeless/tdx/share/qemu/linuxboot_dma.bin";
              path = "/opt/edgeless/${runtime-handler}/tdx/share/qemu/linuxboot_dma.bin";
            }
            {
              url = "file:///opt/edgeless/tdx/share/qemu/efi-virtio.rom";
              path = "/opt/edgeless/${runtime-handler}/tdx/share/qemu/efi-virtio.rom";
            }
          ];
          runtimeHandlerName = runtime-handler;
          inherit (kata.runtime-class-files) debugRuntime;
        };
        destination = "/config/contrast-node-install.json";
      }
    ];
  };

  kata-container-img = ociLayerTar {
    files = [
      {
        source = kata.runtime-class-files.image;
        destination = "/opt/edgeless/share/kata-containers.img";
      }
      {
        source = kata.runtime-class-files.kernel;
        destination = "/opt/edgeless/share/kata-kernel";
      }
    ];
  };

  ovmf-snp = ociLayerTar {
    files = [
      {
        source = kata.runtime-class-files.ovmf-snp;
        destination = "/opt/edgeless/snp/share/OVMF.fd";
      }
    ];
  };

  qemu-snp = ociLayerTar {
    files = [
      {
        source = kata.runtime-class-files.qemu-snp.bin;
        destination = "/opt/edgeless/snp/bin/qemu-system-x86_64";
      }
      {
        source = "${kata.runtime-class-files.qemu-snp.share}/kvmvapic.bin";
        destination = "/opt/edgeless/snp/share/qemu/kvmvapic.bin";
      }
      {
        source = "${kata.runtime-class-files.qemu-snp.share}/linuxboot_dma.bin";
        destination = "/opt/edgeless/snp/share/qemu/linuxboot_dma.bin";
      }
      {
        source = "${kata.runtime-class-files.qemu-snp.share}/efi-virtio.rom";
        destination = "/opt/edgeless/snp/share/qemu/efi-virtio.rom";
      }
    ];
  };

  ovmf-tdx = ociLayerTar {
    files = [
      {
        source = kata.runtime-class-files.ovmf-tdx;
        destination = "/opt/edgeless/tdx/share/OVMF.fd";
      }
    ];
  };

  qemu-tdx = ociLayerTar {
    files = [
      {
        source = kata.runtime-class-files.qemu-tdx.bin;
        destination = "/opt/edgeless/tdx/bin/qemu-system-x86_64";
      }
      {
        source = "${kata.runtime-class-files.qemu-tdx.share}/kvmvapic.bin";
        destination = "/opt/edgeless/tdx/share/qemu/kvmvapic.bin";
      }
      {
        source = "${kata.runtime-class-files.qemu-tdx.share}/linuxboot_dma.bin";
        destination = "/opt/edgeless/tdx/share/qemu/linuxboot_dma.bin";
      }
      {
        source = "${kata.runtime-class-files.qemu-tdx.share}/efi_virtio.rom";
        destination = "/opt/edgeless/tdx/share/qemu/efi-virtio.rom";
      }
    ];
  };

  kata-runtime = ociLayerTar {
    files = [
      {
        source = kata.runtime-class-files.kata-runtime;
        destination = "/opt/edgeless/bin/kata-runtime";
      }
      {
        source = kata.runtime-class-files.containerd-shim-contrast-cc-v2;
        destination = "/opt/edgeless/bin/containerd-shim-contrast-cc-v2";
      }
    ];
  };

  manifest = ociImageManifest {
    layers = [
      node-installer
      installer-config
      kata-container-img
      ovmf-snp
      ovmf-tdx
      qemu-snp
      qemu-tdx
      kata-runtime
    ];
    extraConfig = {
      "config" = {
        "Env" = [
          "PATH=/bin:/usr/bin"
          "CONFIG_DIR=/config"
          "HOST_MOUNT=/host"
        ];
        "Entrypoint" = [ "/bin/node-installer" ];
      };
    };
    extraManifest = {
      "annotations" = {
        "org.opencontainers.image.title" = "contrast-node-installer-kata";
        "org.opencontainers.image.description" = "Contrast Node Installer (Kata)";
        "systems.edgeless.contrast.snp-launch-digest" = launch-digest;
      };
    };
  };
in

ociImageLayout { manifests = [ manifest ]; }
