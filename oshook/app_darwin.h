#import <Cocoa/Cocoa.h>

#ifndef torrent_app_hook_h
#define torrent_app_hook_h

extern void HookDelegate();

extern void OnLoadFileFromPath(char* path);
extern void OnLoadFile(char* data, uint length);
extern void OnLoadURL(char* u);

@interface AppDelegateHooked : NSResponder<NSTouchBarProvider>

@end

@interface Document : NSDocument

@end

@interface URLHandler : NSObject

@end

#endif
