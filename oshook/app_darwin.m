#include "app_darwin.h"

@implementation AppDelegateHooked

// read file runtime
- (BOOL) application:(NSApplication *)sender openFile:(NSString *)filename {
    OnLoadFileFromPath((char*)[filename UTF8String]);
    return YES;
}

@synthesize touchBar;

@end

@interface Document ()

@end

@implementation Document

// read file on init
- (BOOL) readFromData:(NSData *)data ofType:(NSString *)typeName error:(NSError **)outError {
    NSData *dataFromFile = [data retain];
    OnLoadFile((char*)[dataFromFile bytes], (unsigned int)[dataFromFile length]);
    return YES;
}

@end


@implementation URLHandler

// read url
- (void)getURL:(NSAppleEventDescriptor *)event withReplyEvent:(NSAppleEventDescriptor *)reply {
    OnLoadURL((char*)[[[event paramDescriptorForKeyword:keyDirectObject] stringValue] UTF8String]);
}

@end

void HookDelegate() {
	NSApplication* application = [NSApplication sharedApplication];
	[application setDelegate: (id)[[AppDelegateHooked alloc] init]];
    NSLog(@"DELEGATE HOOKED");

    URLHandler* handler = [[URLHandler alloc] init];
    [[NSAppleEventManager sharedAppleEventManager] setEventHandler:handler andSelector:@selector(getURL:withReplyEvent:) forEventClass:kInternetEventClass andEventID:kAEGetURL];
    NSLog(@"URL EVENT HOOKED");
}

