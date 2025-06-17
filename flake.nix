{
  description = "Cross-platform Go project with Gio";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils, ... }:
    let
      goModulePath = "utils/imageFileMover"; # Ensure this matches your go.mod in the src directory
      version = "0.1.0";

      # Inputs for Linux native builds (dev shell and package)
      getLinuxNativeInputs = pkgs: with pkgs; [
        # Cgo compilation
        gcc
        pkg-config

        # Gio graphics stack (EGL/OpenGL/Vulkan)
        libglvnd          # For egl.pc, gl.pc etc.
        mesa.dev          # Mesa drivers and dev files
        vulkan-headers    # For Vulkan API
        vulkan-loader     # For Vulkan ICD loading

        # GTK3 for sqweek/dialog, and GSettings dependencies
        gtk3
        glib              # Core GLib, GSettings functionality, some schemas
        gsettings-desktop-schemas # Standard desktop GSettings schemas

        # Gio X11/Wayland and font dependencies
        xorg.libX11 xorg.libXcursor xorg.libXrandr xorg.libXinerama xorg.libXi
        xorg.libXfixes xorg.libXrender xorg.libXext xorg.libXft
        fontconfig freetype harfbuzz
        wayland libxkbcommon
      ] ++ lib.optionals pkgs.stdenv.isLinux [
        alsa-lib udev
      ];

    in
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        goVersion = pkgs.go;

        # IMPORTANT: Define the source for your Go module.
        # If your go.mod file is in the root of this flake (ImgMover/go.mod):
        #   srcForGoModule = self;
        # If your go.mod file is in a subdirectory like 'src' (ImgMover/src/go.mod):
        srcForGoModule = self + "/src"; # Adjust if your go.mod is at flake root

        nativeInputsList = getLinuxNativeInputs pkgs;

        linuxApp = pkgs.buildGoModule {
          pname = "imgmover";
          version = version;
          src = srcForGoModule; # Use the correct source path
          inherit goModulePath;

          vendorSha256 = pkgs.lib.fakeSha256; # Replace with actual hash after first build

          buildInputs = [ goVersion ] ++ nativeInputsList ++ [
            pkgs.dconf # dconf GSettings backend needed at runtime
          ];

          # For packaged GUI applications, wrapGAppsHook sets up GSettings, XDG paths, etc.
          nativeBuildInputs = [
            pkgs.makeWrapper
            pkgs.wrapGAppsHook
          ];

          # wrapGAppsHook should handle GSETTINGS_SCHEMA_DIR and XDG_DATA_DIRS.
          # We might need to explicitly add GIO_EXTRA_MODULES for dconf.
          # Check if wrapGAppsHook is enough; if not, add/uncomment postFixup.
           postFixup = ''
             wrapProgram $out/bin/${goModulePath} \
               --prefix GIO_EXTRA_MODULES : "${pkgs.dconf}/lib/gio/modules:${pkgs.glib-networking}/lib/gio/modules"
          #    # If you need to ensure specific XDG_DATA_DIRS or GSETTINGS_SCHEMA_DIR beyond what wrapGAppsHook provides:
          #    # --prefix XDG_DATA_DIRS : "${pkgs.gtk3}/share:${pkgs.gsettings-desktop-schemas}/share" \
          #    # --prefix GSETTINGS_SCHEMA_DIR : "${pkgs.glib}/share/gsettings-schemas/${pkgs.glib.name}/glib-2.0/schemas:${pkgs.gsettings-desktop-schemas}/share/gsettings-schemas/${pkgs.gsettings-desktop-schemas.name}/glib-2.0/schemas:${pkgs.gtk3}/share/gsettings-schemas/${pkgs.gtk3.name}/glib-2.0/schemas"
           '';
        };

        pkgsWindows = nixpkgs.legacyPackages.${system}.pkgsCross.mingwW64;
        windowsApp = pkgsWindows.buildGoModule {
          pname = "imgmover-windows";
          version = version;
          src = srcForGoModule; # Use the correct source path
          inherit goModulePath;
          vendorSha256 = pkgs.lib.fakeSha256;
          buildInputs = [
            # Go compiler comes from host
          ];
          postInstall = ''
            mv $out/bin/${goModulePath} $out/bin/imgmover.exe
          '';
        };

      in
      {
        packages = {
          default = linuxApp;
          imgmover-linux = linuxApp;
          imgmover-windows = windowsApp;
        };

        devShells.default = pkgs.mkShell {
          name = "go-gio-dev-env";
          packages = [
            goVersion
            pkgs.fish
            pkgs.git
            # Runtime dependencies for GSettings/GTK in dev shell
            pkgs.dconf                # GSettings backend
            pkgs.glib-networking      # For GIO networking modules
            pkgs.shared-mime-info   # For XDG_DATA_DIRS content
          ] ++ nativeInputsList; # Includes gtk3, glib, gsettings-desktop-schemas etc.

          shellHook = ''
            export GOROOT="${goVersion}/share/go"
            echo "Welcome to the Go-Gio (Fish) development shell for NixOS/Wayland!"
            echo "Go version: $(go version)"

            # --- Explicit GSettings and XDG Setup for the Dev Shell ---
            # This is often necessary because mkShell doesn't always set them up
            # as comprehensively as wrapGAppsHook does for packaged apps.

            SCHEMA_DIRS_TEMP=""
            # Packages known to provide GSettings schemas
            declare -a SCHEMA_PACKAGES_TEMP=("${pkgs.glib}" "${pkgs.gsettings-desktop-schemas}" "${pkgs.gtk3}")
            for pkg_path_temp in "''${SCHEMA_PACKAGES_TEMP[@]}"; do
              # Try to find the actual schema subdirectory (name can vary with version)
              # e.g., glib-2.0, gtk-3.0 etc. inside the package's .../gsettings-schemas/
              actual_schema_root_dir=$(find "$pkg_path_temp/share/gsettings-schemas/" -maxdepth 1 -mindepth 1 -type d -print -quit 2>/dev/null)
              if [ -n "$actual_schema_root_dir" ] && [ -d "$actual_schema_root_dir/glib-2.0/schemas" ]; then
                SCHEMA_DIRS_TEMP="''${SCHEMA_DIRS_TEMP:+$SCHEMA_DIRS_TEMP:}$actual_schema_root_dir/glib-2.0/schemas"
              # Fallback for older structures or if the above find fails
              elif [ -d "$pkg_path_temp/share/gsettings-schemas/glib-2.0/schemas" ]; then
                 SCHEMA_DIRS_TEMP="''${SCHEMA_DIRS_TEMP:+$SCHEMA_DIRS_TEMP:}$pkg_path_temp/share/gsettings-schemas/glib-2.0/schemas"
              fi
            done
            export GSETTINGS_SCHEMA_DIR="''${GSETTINGS_SCHEMA_DIR:+$GSETTINGS_SCHEMA_DIR:}$SCHEMA_DIRS_TEMP"

            DATA_DIRS_TEMP=""
            # Packages providing XDG data (icons, themes, mime info)
            declare -a DATA_PACKAGES_TEMP=("${pkgs.gtk3}" "${pkgs.gsettings-desktop-schemas}" "${pkgs.shared-mime-info}")
            for pkg_path_temp in "''${DATA_PACKAGES_TEMP[@]}"; do
              if [ -d "$pkg_path_temp/share" ]; then
                DATA_DIRS_TEMP="''${DATA_DIRS_TEMP:+$DATA_DIRS_TEMP:}$pkg_path_temp/share"
              fi
            done
            export XDG_DATA_DIRS="''${XDG_DATA_DIRS:+$XDG_DATA_DIRS:}$DATA_DIRS_TEMP"

            # GIO_EXTRA_MODULES for dconf GSettings backend and networking
            export GIO_EXTRA_MODULES="''${GIO_EXTRA_MODULES:+$GIO_EXTRA_MODULES:}${pkgs.dconf}/lib/gio/modules:${pkgs.glib-networking}/lib/gio/modules"

            echo "Dev GSETTINGS_SCHEMA_DIR=$GSETTINGS_SCHEMA_DIR"
            echo "Dev XDG_DATA_DIRS=$XDG_DATA_DIRS"
            echo "Dev GIO_EXTRA_MODULES=$GIO_EXTRA_MODULES"
            # --- End GSettings Setup ---

            if [ -z "$FISH_VERSION" ]; then
              echo "Switching to Fish shell..."
              exec fish
            fi
          '';
        };
      }
    ) // {
      checks = flake-utils.lib.eachDefaultSystem (system:
          let
            pkgs = nixpkgs.legacyPackages.${system};
            in
            {
            default = self.outputs.packages.${system}.default;
            windows = if pkgs.stdenv.isLinux then self.outputs.packages.${system}.imgmover-windows else null;
        }); };
}
