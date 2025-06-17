{
  description = "Cross-platform Go project with Gio";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils, ... }:
    let
      # Common Go module info - replace with your actual module path
      goModulePath = "example.com/imgmover"; # e.g., gitlab.com/youruser/imgmover
      version = "0.1.0"; # Your app's version

      # Helper to get gio build inputs for Linux
      getLinuxGioBuildInputs = pkgs: with pkgs; [
        gcc # Native C compiler for Cgo
        pkg-config
        xorg.libX11 xorg.libXcursor xorg.libXrandr xorg.libXinerama xorg.libXi
        xorg.libXfixes xorg.libXrender xorg.libXext xorg.libXft
        fontconfig freetype harfbuzz
        wayland libxkbcommon mesa
      ] ++ lib.optionals pkgs.stdenv.isLinux [
        alsa-lib udev
      ];

    in
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        goVersion = pkgs.go_1_22; # Or your preferred Go version

        # Native Linux build
        linuxApp = pkgs.buildGoModule {
          pname = "imgmover";
          inherit version src goModulePath;

          # For reproducible builds, vendor your dependencies:
          # 1. Run `go mod tidy`
          # 2. Run `go mod vendor`
          # 3. Calculate the hash: `nix-hash --type sha256 --base32 ./vendor`
          # vendorHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="; # Replace with actual hash
          # Or, if you prefer Nix to fetch:
          # vendorSha256 = pkgs.lib.fakeSha256; # For initial testing, then get the real one from build failure

          # CGO_ENABLED=1 is required for Gio and is usually default if buildInputs are present
          buildInputs = [ goVersion ] ++ (getLinuxGioBuildInputs pkgs);

          # Ensure the output binary has a predictable name if needed
          # postInstall = ''
          #  mv $out/bin/${pname} $out/bin/imgmover-linux
          # '';
        };

        # Windows 64-bit cross-compilation
        # We define this regardless of the host `system` because we're *targeting* Windows.
        # The actual build will happen on a Linux host if `nix build .#imgmover-windows` is run there.
        pkgsWindows = nixpkgs.legacyPackages.${system}.pkgsCross.mingwW64; # MinGW 64-bit toolchain
        windowsApp = pkgsWindows.buildGoModule {
          pname = "imgmover-windows";
          inherit version src goModulePath;

          # vendorHash or vendorSha256 (can often be the same as for Linux if go.sum is platform-agnostic)
          # vendorHash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA=";

          # buildGoModule for a cross target automatically sets GOOS and GOARCH.
          # CGO_ENABLED=1 is still needed.
          # The MinGW toolchain (gcc for windows) is implicitly part of pkgsWindows.stdenv.cc
          # Gio for Windows links against system DLLs (user32, gdi32, etc.),
          # which the MinGW toolchain provides import libraries for.
          # It typically does NOT need X11, Wayland, fontconfig, etc.
          # If Gio bundles its own version of FreeType for Windows, you might need pkgsWindows.freetype.
          buildInputs = [
            # goVersion # The Go compiler itself comes from the *host* pkgs for cross-compilation
                        # buildGoModule from pkgsWindows should handle this correctly.
                        # If you face issues, you might need to specify `go = goVersion;`
          ];

          # Go build tags specific to Gio for Windows might be needed if Gio uses them.
          # buildFlags = [ "-tags=nowayland" ]; # Example, check Gio docs if needed

          # The output binary will be .exe
          postInstall = ''
            mv $out/bin/${goModulePath} $out/bin/imgmover.exe
          '';
        };

      in
      {
        # Packages for `nix build`
        packages = {
          default = linuxApp; # `nix build` will build the Linux version by default
          imgmover-linux = linuxApp;
          imgmover-windows = windowsApp;
        };

        # Development shell (for Linux development)
        devShells.default = pkgs.mkShell {
          name = "go-gio-dev-env";
          packages = [
            goVersion
            pkgs.fish
            pkgs.git
            # pkgs.delve # Go debugger
            # pkgs.gopls # Go language server
          ] ++ (getLinuxGioBuildInputs pkgs);

          shellHook = ''
            export GOROOT="${goVersion}/share/go"
            echo "Welcome to the Go-Gio (Fish) development shell!"
            echo "Go version: $(go version)"

            if [ -z "$FISH_VERSION" ]; then
              echo "Switching to Fish shell..."
              exec fish
            fi
          '';
        };
      }
    ) // { # Add outputs not dependent on flake-utils.eachDefaultSystem if needed
      # Example: A flake check that builds both
      checks = flake-utils.lib.eachDefaultSystem (system: {
        default = self.outputs.packages.${system}.default;
        windows = self.outputs.packages.${system}.imgmover-windows; # This assumes you can cross-compile from `system`
      });
    };
}
