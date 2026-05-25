declare module '../../wailsjs/go/main/App' {
  export function ReadCodexConfigToml(): Promise<string>
  export function RestoreCodexConfigToml(): Promise<string>
  export function WriteCodexConfigTomlRaw(content: string): Promise<string>
}
