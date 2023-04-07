#import <Cocoa/Cocoa.h>

#ifndef torrent_app_hook_h
#define torrent_app_hook_h

extern void HookDelegate();

@interface AppDelegateHooked : NSResponder<NSTouchBarProvider>

@end

@interface Document : NSDocument

@end

@interface URLHandler : NSObject

@end

#endif
