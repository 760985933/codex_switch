import sys

with open('/Users/tianfeng/Documents/codes/codex接入deepseek/frontend/wailsjs/go/models.ts', 'r') as f:
    content = f.read()

# Add provider field to Profile class
old = '''		    name: string;
		    baseURL: string;'''
new = '''		    name: string;
		    provider: string;
		    baseURL: string;'''
content = content.replace(old, new)

# Add provider assignment in constructor
old2 = '''this.name = source["name"];
		        this.baseURL = source["baseURL"];'''
new2 = '''this.name = source["name"];
		        this.provider = source["provider"];
		        this.baseURL = source["baseURL"];'''
content = content.replace(old2, new2)

with open('/Users/tianfeng/Documents/codes/codex接入deepseek/frontend/wailsjs/go/models.ts', 'w') as f:
    f.write(content)

print("Done")
