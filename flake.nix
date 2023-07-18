{
  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs";
  };

  outputs = {nixpkgs, ...}: let
    inherit (nixpkgs) lib;
    genSystems = lib.genAttrs [
      "x86_64-linux"
      "aarch64-linux"
      "x86_64-darwin"
      "aarch64-darwin"
    ];

    nixpkgsFor = system: import nixpkgs {inherit system;};

    genWithPkgs = f: genSystems (system: f (nixpkgsFor system));
  in {
    packages = genWithPkgs (pkgs: rec {
      ssg = pkgs.buildGoModule {
        pname = "ssg";
        version = "0.1.0";

        src = ./.;

        vendorHash = "sha256-YjMxgze5Yf178rOLwj4ctRL+XXyCgGd9TGf8gnWLhKQ=";
      };
      buildSite = { contentDir, staticDir, templateDir }: pkgs.stdenv.mkDerivation {
        name = "ssg-site";

        inherit contentDir staticDir templateDir;

        nativeBuildInputs = [ ssg ];

        dontUnpack = true; # we don't have sources; unpacking will cause an error

        buildPhase = ''
          site -contentDir $contentDir -staticDir $staticDir -templateDir $templateDir -out $out
        '';
      };
      lunaSite = buildSite {
        contentDir = ./content;
        staticDir = ./static;
        templateDir = ./templates;
      };
    });
  
    devShells = genWithPkgs (pkgs: {
      default = pkgs.mkShell {
        packages = with pkgs; [
          go
          gopls
          delve
        ];
      };
    });
    
    formatter = genWithPkgs (pkgs: pkgs.alejandra);
  };
}
