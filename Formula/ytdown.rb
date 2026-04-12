cask "ytdown" do
  version :latest

  url do |page|
    require "net/http"
    require "json"

    # Fetch the latest release from GitHub API
    uri = URI("https://api.github.com/repos/JustinNguyen9979/YTDown/releases/latest")
    http = Net::HTTP.new(uri.host, uri.port)
    http.use_ssl = true

    request = Net::HTTP::Get.new(uri.path)
    request["Accept"] = "application/vnd.github+json"

    response = http.request(request)
    if response.code == "200"
      release_data = JSON.parse(response.body)
      
      # Find the DMG asset
      dmg_asset = release_data["assets"].find { |asset| asset["name"].end_with?(".dmg") }
      raise "No DMG found in latest release" unless dmg_asset
      
      dmg_asset["browser_download_url"]
    else
      raise "Failed to fetch latest release from GitHub"
    end
  end

  appcast "https://api.github.com/repos/JustinNguyen9979/YTDown/releases.atom"

  name "YTDown"
  desc "YouTube video downloader for macOS"
  homepage "https://github.com/JustinNguyen9979/YTDown"

  depends_on macos: ">= :big_sur"

  app "YTDown.app"

  uninstall quit: "dev.ytdown.app"

  zap trash: [
    "~/Library/Application Support/YTDown",
    "~/Library/Caches/dev.ytdown.app",
    "~/Library/Preferences/dev.ytdown.app.plist",
  ]
end
