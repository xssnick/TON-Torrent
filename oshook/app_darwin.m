#include "app_darwin.h"

@implementation AppDelegateHooked

-(BOOL)application:(NSApplication *)sender openFile:(NSString *)filename {
    OnLoadFileFromPath([filename UTF8String]);
    return YES;
}

@synthesize touchBar;

@end

@interface Document ()

@end

@implementation Document

- (BOOL) readFromData:(NSData *)data ofType:(NSString *)typeName error:(NSError **)outError {
    NSData *dataFromFile = [data retain];
    OnLoadFile([dataFromFile bytes], (unsigned int)[dataFromFile length]);
    return YES;
}

@end


void HookDelegate() {
	NSApplication* application = [NSApplication sharedApplication];
	[application setDelegate: (id)[[AppDelegateHooked alloc] init]];
    NSLog(@"DELEGATE HOOKED");
}