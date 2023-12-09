Unicode true

####
## Please note: Template replacements don't work in this file. They are provided with default defines like
## mentioned underneath.
## If the keyword is not defined, "wails_tools.nsh" will populate them with the values from ProjectInfo. 
## If they are defined here, "wails_tools.nsh" will not touch them. This allows to use this project.nsi manually 
## from outside of Wails for debugging and development of the installer.
## 
## For development first make a wails nsis build to populate the "wails_tools.nsh":
## > wails build --target windows/amd64 --nsis
## Then you can call makensis on this file with specifying the path to your binary:
## For a AMD64 only installer:
## > makensis -DARG_WAILS_AMD64_BINARY=..\..\bin\app.exe
## For a ARM64 only installer:
## > makensis -DARG_WAILS_ARM64_BINARY=..\..\bin\app.exe
## For a installer with both architectures:
## > makensis -DARG_WAILS_AMD64_BINARY=..\..\bin\app-amd64.exe -DARG_WAILS_ARM64_BINARY=..\..\bin\app-arm64.exe
####
## The following information is taken from the ProjectInfo file, but they can be overwritten here. 
####
## !define INFO_PROJECTNAME    "MyProject" # Default "{{.Name}}"
## !define INFO_COMPANYNAME    "MyCompany" # Default "{{.Info.CompanyName}}"
## !define INFO_PRODUCTNAME    "MyProduct" # Default "{{.Info.ProductName}}"
## !define INFO_PRODUCTVERSION "1.1.0"     # Default "{{.Info.ProductVersion}}"
## !define INFO_COPYRIGHT      "Copyright" # Default "{{.Info.Copyright}}"
###
## !define PRODUCT_EXECUTABLE  "Application.exe"      # Default "${INFO_PROJECTNAME}.exe"
## !define UNINST_KEY_NAME     "UninstKeyInRegistry"  # Default "${INFO_COMPANYNAME}${INFO_PRODUCTNAME}"
####
## !define REQUEST_EXECUTION_LEVEL "admin"            # Default "admin"  see also https://nsis.sourceforge.io/Docs/Chapter4.html
####
## Include the wails tools
####
!include "wails_tools.nsh"

# The version information for this two must consist of 4 parts
VIProductVersion "${INFO_PRODUCTVERSION}.0"
VIFileVersion    "${INFO_PRODUCTVERSION}.0"

VIAddVersionKey "CompanyName"     "${INFO_COMPANYNAME}"
VIAddVersionKey "FileDescription" "${INFO_PRODUCTNAME} Installer"
VIAddVersionKey "ProductVersion"  "${INFO_PRODUCTVERSION}"
VIAddVersionKey "FileVersion"     "${INFO_PRODUCTVERSION}"
VIAddVersionKey "LegalCopyright"  "${INFO_COPYRIGHT}"
VIAddVersionKey "ProductName"     "${INFO_PRODUCTNAME}"

!include "MUI.nsh"

!define MUI_ICON "..\icon.ico"
!define MUI_UNICON "..\icon.ico"
# !define MUI_WELCOMEFINISHPAGE_BITMAP "resources\leftimage.bmp" #Include this to add a bitmap on the left side of the Welcome Page. Must be a size of 164x314
!define MUI_FINISHPAGE_NOAUTOCLOSE # Wait on the INSTFILES page so the user can take a look into the details of the installation steps
!define MUI_ABORTWARNING # This will warn the user if they exit from the installer.
!define MUI_FINISHPAGE_RUN "$INSTDIR\${PRODUCT_EXECUTABLE}"

!insertmacro MUI_PAGE_WELCOME # Welcome to the installer page.
# !insertmacro MUI_PAGE_LICENSE "resources\eula.txt" # Adds a EULA page to the installer
!insertmacro MUI_PAGE_DIRECTORY # In which folder install page.
!insertmacro MUI_PAGE_INSTFILES # Installing page.
!insertmacro MUI_PAGE_FINISH # Finished installation page.

!insertmacro MUI_UNPAGE_INSTFILES # Uinstalling page

!insertmacro MUI_LANGUAGE "English" # Set the Language of the installer

## The following two statements can be used to sign the installer and the uninstaller. The path to the binaries are provided in %1
#!uninstfinalize 'signtool --file "%1"'
#!finalize 'signtool --file "%1"'

Name "${INFO_PRODUCTNAME}"
OutFile "..\..\bin\${INFO_PROJECTNAME}-${ARCH}-installer.exe" # Name of the installer's file.
InstallDir "$PROGRAMFILES64\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}" # Default installing folder ($PROGRAMFILES is Program Files folder).
ShowInstDetails show # This will always show the installation details.

Function .onInit
   !insertmacro wails.checkArchitecture
FunctionEnd

Section
    !insertmacro wails.webview2runtime

    SetOutPath $INSTDIR
    
    !insertmacro wails.files
    CreateShortcut "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"
    CreateShortCut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"

    !insertmacro wails.writeUninstaller
SectionEnd

!define ASSOC_EXT ".tonbag"
!define ASSOC_URL1 "tonstorage"
!define ASSOC_URL2 "tonbag"
!define ASSOC_PROGID "TON Torrent"
!define ASSOC_VERB "open"
Section -ShellAssoc
  # Register file type and url
  WriteRegStr ShCtx "Software\Classes\${ASSOC_PROGID}\DefaultIcon" "" "$InstDir\${PRODUCT_EXECUTABLE},0"
  WriteRegStr ShCtx "Software\Classes\${ASSOC_PROGID}\shell\${ASSOC_VERB}\command" "" '"$InstDir\${PRODUCT_EXECUTABLE}" "%1"'
  WriteRegStr ShCtx "Software\Classes\${ASSOC_EXT}" "" "${ASSOC_PROGID}"

  WriteRegStr ShCtx "Software\Classes\${ASSOC_URL1}" "" "${ASSOC_PROGID}"
  WriteRegStr ShCtx "Software\Classes\${ASSOC_URL1}" "URL Protocol" ""
  WriteRegStr ShCtx "Software\Classes\${ASSOC_URL1}" "Content Type" "application/ton-torrent"
  WriteRegStr ShCtx "Software\Classes\${ASSOC_URL1}\DefaultIcon" "" "$InstDir\${PRODUCT_EXECUTABLE},0"
  WriteRegStr ShCtx "Software\Classes\${ASSOC_URL1}\shell\${ASSOC_VERB}\command" "" '"$InstDir\${PRODUCT_EXECUTABLE}" "%1"'

  WriteRegStr ShCtx "Software\Classes\${ASSOC_URL2}" "" "${ASSOC_PROGID}"
  WriteRegStr ShCtx "Software\Classes\${ASSOC_URL2}" "URL Protocol" ""
  WriteRegStr ShCtx "Software\Classes\${ASSOC_URL2}" "Content Type" "application/ton-torrent"
  WriteRegStr ShCtx "Software\Classes\${ASSOC_URL2}\DefaultIcon" "" "$InstDir\${PRODUCT_EXECUTABLE},0"
  WriteRegStr ShCtx "Software\Classes\${ASSOC_URL2}\shell\${ASSOC_VERB}\command" "" '"$InstDir\${PRODUCT_EXECUTABLE}" "%1"'

  # Register "Open With" [Optional]
  WriteRegNone ShCtx "Software\Classes\${ASSOC_EXT}\OpenWithList" "${PRODUCT_EXECUTABLE}" ; Win2000+ [Optional]
  WriteRegStr ShCtx "Software\Classes\Applications\${PRODUCT_EXECUTABLE}\shell\open\command" "" '"$InstDir\${PRODUCT_EXECUTABLE}" "%1"'
  WriteRegStr ShCtx "Software\Classes\Applications\${PRODUCT_EXECUTABLE}" "FriendlyAppName" "TON Torrent" ; [Optional]
  WriteRegStr ShCtx "Software\Classes\Applications\${PRODUCT_EXECUTABLE}" "ApplicationCompany" "Tonutils" ; [Optional]
  WriteRegNone ShCtx "Software\Classes\Applications\${PRODUCT_EXECUTABLE}\SupportedTypes" "${ASSOC_EXT}" ; [Optional] Only allow "Open With" with specific extension(s) on WinXP+

  # Register "Default Programs" [Optional]
  !ifdef REGISTER_DEFAULTPROGRAMS
  WriteRegStr ShCtx "Software\Classes\Applications\${PRODUCT_EXECUTABLE}\Capabilities" "ApplicationDescription" "TON Storage torrent client"
  WriteRegStr ShCtx "Software\Classes\Applications\${PRODUCT_EXECUTABLE}\Capabilities\FileAssociations" "${ASSOC_EXT}" "${ASSOC_PROGID}"
  WriteRegStr ShCtx "Software\RegisteredApplications" "TON Torrent" "Software\Classes\Applications\${PRODUCT_EXECUTABLE}\Capabilities"
  !endif
SectionEnd

Section -un.ShellAssoc
  # Unregister file type
  ClearErrors
  DeleteRegKey ShCtx "Software\Classes\${ASSOC_PROGID}\shell\${ASSOC_VERB}"
  DeleteRegKey /IfEmpty ShCtx "Software\Classes\${ASSOC_PROGID}\shell"
  ${IfNot} ${Errors}
    DeleteRegKey ShCtx "Software\Classes\${ASSOC_PROGID}\DefaultIcon"
  ${EndIf}
  ReadRegStr $0 ShCtx "Software\Classes\${ASSOC_EXT}" ""
  DeleteRegKey /IfEmpty ShCtx "Software\Classes\${ASSOC_PROGID}"
  ${IfNot} ${Errors}
  ${AndIf} $0 == "${ASSOC_PROGID}"
    DeleteRegValue ShCtx "Software\Classes\${ASSOC_EXT}" ""
    DeleteRegKey /IfEmpty ShCtx "Software\Classes\${ASSOC_EXT}"
  ${EndIf}

  ReadRegStr $0 ShCtx "Software\Classes\${ASSOC_URL1}" ""
  DeleteRegKey ShCtx "Software\Classes\${ASSOC_URL1}"
  ${IfNot} ${Errors}
  ${AndIf} $0 == "${ASSOC_PROGID}"
    DeleteRegKey ShCtx "Software\Classes\${ASSOC_URL1}"
  ${EndIf}

  ReadRegStr $0 ShCtx "Software\Classes\${ASSOC_URL2}" ""
  DeleteRegKey ShCtx "Software\Classes\${ASSOC_URL2}"
  ${IfNot} ${Errors}
  ${AndIf} $0 == "${ASSOC_PROGID}"
    DeleteRegKey ShCtx "Software\Classes\${ASSOC_URL2}"
  ${EndIf}

  # Unregister "Open With"
  DeleteRegKey ShCtx "Software\Classes\Applications\${PRODUCT_EXECUTABLE}"
  DeleteRegValue ShCtx "Software\Classes\${ASSOC_EXT}\OpenWithList" "${PRODUCT_EXECUTABLE}"
  DeleteRegKey /IfEmpty ShCtx "Software\Classes\${ASSOC_EXT}\OpenWithList"
  DeleteRegValue ShCtx "Software\Classes\${ASSOC_EXT}\OpenWithProgids" "${ASSOC_PROGID}"
  DeleteRegKey /IfEmpty ShCtx "Software\Classes\${ASSOC_EXT}\OpenWithProgids"
  DeleteRegKey /IfEmpty  ShCtx "Software\Classes\${ASSOC_EXT}"

  # Unregister "Default Programs"
  !ifdef REGISTER_DEFAULTPROGRAMS
  DeleteRegValue ShCtx "Software\RegisteredApplications" "TON Torrent"
  DeleteRegKey ShCtx "Software\Classes\Applications\${PRODUCT_EXECUTABLE}\Capabilities"
  DeleteRegKey /IfEmpty ShCtx "Software\Classes\Applications\${PRODUCT_EXECUTABLE}"
  !endif

  # Attempt to clean up junk left behind by the Windows shell
  DeleteRegValue HKCU "Software\Microsoft\Windows\CurrentVersion\Search\JumplistData" "$InstDir\${PRODUCT_EXECUTABLE}"
  DeleteRegValue HKCU "Software\Classes\Local Settings\Software\Microsoft\Windows\Shell\MuiCache" "$InstDir\${PRODUCT_EXECUTABLE}.FriendlyAppName"
  DeleteRegValue HKCU "Software\Classes\Local Settings\Software\Microsoft\Windows\Shell\MuiCache" "$InstDir\${PRODUCT_EXECUTABLE}.ApplicationCompany"
  DeleteRegValue HKCU "Software\Microsoft\Windows\ShellNoRoam\MUICache" "$InstDir\${PRODUCT_EXECUTABLE}" ; WinXP
  DeleteRegValue HKCU "Software\Microsoft\Windows NT\CurrentVersion\AppCompatFlags\Compatibility Assistant\Store" "$InstDir\${PRODUCT_EXECUTABLE}"
  DeleteRegValue HKCU "Software\Microsoft\Windows\CurrentVersion\ApplicationAssociationToasts" "${ASSOC_PROGID}_${ASSOC_EXT}"
  DeleteRegValue HKCU "Software\Microsoft\Windows\CurrentVersion\ApplicationAssociationToasts" "Applications\${PRODUCT_EXECUTABLE}_${ASSOC_EXT}"
  DeleteRegValue HKCU "Software\Microsoft\Windows\CurrentVersion\Explorer\FileExts\${ASSOC_EXT}\OpenWithProgids" "${ASSOC_PROGID}"
  DeleteRegKey /IfEmpty HKCU "Software\Microsoft\Windows\CurrentVersion\Explorer\FileExts\${ASSOC_EXT}\OpenWithProgids"
  DeleteRegKey /IfEmpty HKCU "Software\Microsoft\Windows\CurrentVersion\Explorer\FileExts\${ASSOC_EXT}\OpenWithList"
  DeleteRegKey /IfEmpty HKCU "Software\Microsoft\Windows\CurrentVersion\Explorer\FileExts\${ASSOC_EXT}"
  ;DeleteRegKey HKCU "Software\Microsoft\Windows\Roaming\OpenWith\FileExts\${ASSOC_EXT}"
  ;DeleteRegKey HKCU "Software\Microsoft\Windows\CurrentVersion\Explorer\RecentDocs\${ASSOC_EXT}"

SectionEnd

Section "uninstall" 
    RMDir /r "$AppData\${PRODUCT_EXECUTABLE}" # Remove the WebView2 DataPath

    RMDir /r $INSTDIR

    Delete "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk"
    Delete "$DESKTOP\${INFO_PRODUCTNAME}.lnk"

    !insertmacro wails.deleteUninstaller
SectionEnd
