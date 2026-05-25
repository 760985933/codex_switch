declare module '../../wailsjs/go/main/App' {
  export function ClearCodexConfigBackups(): Promise<number>
  export function DeleteCodexConfigBackup(backupPath: string): Promise<string>
  export function ListCodexConfigBackups(): Promise<string[]>
  export function ReadCodexConfigToml(): Promise<string>
  export function RestoreCodexConfigToml(): Promise<string>
  export function RestoreCodexConfigTomlFromBackup(backupPath: string): Promise<string>
  export function WriteCodexConfigTomlRaw(content: string): Promise<string>
}
