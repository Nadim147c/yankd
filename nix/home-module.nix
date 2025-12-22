self:
{
  config,
  lib,
  pkgs,
  ...
}:
let
  inherit (lib)
    getExe
    literalExpression
    mkEnableOption
    mkIf
    mkOption
    ;
  inherit (lib.types)
    attrs
    listOf
    package
    str
    ;

  pkg = self.packages.${pkgs.stdenv.hostPlatform.system}.default;
  cfg = config.services.yankd;
  format = pkgs.formats.json { };
in
{
  options.services.yankd = {
    enable = mkEnableOption "Enable yankd clipboard manager";

    package = mkOption {
      type = package;
      default = pkg;
      description = "Yankd package to use";
    };

    systemdTargets = mkOption {
      type = listOf str;
      default = [ config.wayland.systemd.target ];
      defaultText = literalExpression "[ config.wayland.systemd.target ]";
      example = "wayland-session@Hyprland.target";
      description = "The systemd targets that will automatically start the yankd service.";
    };

    settings = mkOption {
      type = attrs;
      default = { };
      description = "Yankd settings to use";
    };
  };

  config = mkIf cfg.enable {
    home.packages = mkIf (cfg.package != null) [ cfg.package ];

    xdg.configFile = mkIf (cfg.settings != { }) {
      "yankd/config.json".source = format.generate "yankd.json" cfg.settings;
    };

    systemd.user.services.yankd = {
      Unit = {
        Description = "Clipboard management daemon";
        PartOf = cfg.systemdTargets;
        After = cfg.systemdTargets;
      };

      Service = {
        Type = "simple";
        ExecStart = "${getExe cfg.package} watch -v";
        Restart = "on-failure";
      };

      Install = {
        WantedBy = cfg.systemdTargets;
      };
    };
  };
}
