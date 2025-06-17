{
  description = "Cross-platform Go project with Gio";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils, ... }: # 'self' here refers to the flake itself
    let
      goModulePath = "utils/imageFileMover";
      version = "0.1.0";

      getLinuxGioBuildInputs = pkgs: with pkgs; [
        gcc pkg-config
        xorg.libX11 xorg.libXcursor xorg.libXrandr xorg.libXinerama xorg.libXi
        xorg.libXfixes xorg.libXrender xorg.libXext xorg.libXft
        fontconfig freetype harfbuzz
        wayland libxkbcommon mesa libglvnd
        vulkan-headers vulkan-loader gtk3
      ] ++ lib.optionals pkgs.stdenv.isLinux [
        alsa-lib udev
      ];

    in
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        goVersion = pkgs.go;

        # Define src here to be the flake's own source files
        src = self; 
        linuxApp = pkgs.buildGoModule {
          pname = "imgmover";
          inherit version src goModulePath; # Now 'src' is defined in this scope

          vendorSha256 = pkgs.lib.fakeSha256; # For initial build, then replace

          buildInputs = [ goVersion ] ++ (getLinuxGioBuildInputs pkgs);
        };

        pkgsWindows = nixpkgs.legacyPackages.${system}.pkgsCross.mingwW64;
        windowsApp = pkgsWindows.buildGoModule {
          pname = "imgmover-windows";
          inherit version src goModulePath; # 'src' is also available here

          vendorSha256 = pkgs.lib.fakeSha256; # For initial build, then replace

          buildInputs = [
            # Go compiler comes from host for cross-compilation
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
          ] ++ (getLinuxGioBuildInputs pkgs);

          shellHook = ''
            export GOROOT="${goVersion}/share/go"
            echo "Welcome to the Go-Gio (Fish) development shell for NixOS/Wayland!"
            echo "Go version: $(go version)"

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
        # Only try to build windows if the host system is Linux, as cross-compiling
        # from Darwin to Windows with MinGW in Nixpkgs can be more complex or unsupported.
        # You can adjust this condition if needed.
            windows = if pkgs.stdenv.isLinux then self.outputs.packages.${system}.imgmover-windows else null;
        }); };
}
