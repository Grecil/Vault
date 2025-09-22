import React from 'react'
import { useUser } from '@clerk/clerk-react'
import { VaultIcon } from './FileTypeIcons'
import { useStorageStats } from '../hooks/useStorageStats'

export interface SidebarItem {
  id: string
  name: string
  icon: React.ReactNode
}

interface DashboardSidebarProps {
  sidebarItems: SidebarItem[]
  activeTab: string
  onTabChange: (tabId: string) => void
  fileCount: number
}

const DashboardSidebar: React.FC<DashboardSidebarProps> = ({
  sidebarItems,
  activeTab,
  onTabChange,
  fileCount
}) => {
  const { user } = useUser()
  const { statistics, loading: statsLoading, formatBytes } = useStorageStats()

  return (
    <div className="w-64 bg-sidebar border-r border-sidebar-border flex flex-col h-full">
      {/* Logo */}
      <div className="p-6 border-b border-sidebar-border">
        <div className="flex items-center space-x-2">
          <div className="w-8 h-8 bg-sidebar-primary rounded-lg flex items-center justify-center">
            <VaultIcon/>
          </div>
          <span className="text-xl font-bold text-sidebar-foreground">Vault</span>
        </div>
      </div>

      {/* User Info */}
      <div className="p-4 border-b border-sidebar-border">
        <div className="flex items-center space-x-3">
          <div className="w-10 h-10 bg-sidebar-primary rounded-full flex items-center justify-center">
            <span className="text-sidebar-primary-foreground font-medium">
              {user?.firstName?.charAt(0) || 'U'}
            </span>
          </div>
          <div>
            <p className="text-sm font-medium text-sidebar-foreground">
              {user?.firstName || 'User'}
            </p>
            <p className="text-xs text-muted-foreground">
              {fileCount} files
            </p>
          </div>
        </div>
      </div>

      {/* Navigation */}
      <nav className="flex-1 p-4">
        <ul className="space-y-1">
          {sidebarItems.map((item) => (
            <li key={item.id}>
              <button
                onClick={() => onTabChange(item.id)}
                className={`w-full flex items-center space-x-3 px-3 py-2 rounded-lg text-left transition-colors ${
                  activeTab === item.id
                    ? 'bg-sidebar-accent text-sidebar-primary'
                    : 'text-sidebar-foreground hover:bg-sidebar-accent/50'
                }`}
              >
                {item.icon}
                <span className="font-medium">{item.name}</span>
              </button>
            </li>
          ))}
        </ul>
      </nav>

      {/* Storage Info - Stick to bottom */}
      <div className="mt-auto p-4 border-t border-sidebar-border">
        <div className="space-y-2">
          {statsLoading ? (
            <div className="animate-pulse">
              <div className="flex justify-between text-xs">
                <span className="text-muted-foreground">Loading...</span>
                <div className="h-3 bg-muted rounded w-16"></div>
              </div>
              <div className="w-full bg-muted rounded-full h-2 mt-1">
                <div className="bg-muted h-2 rounded-full"></div>
              </div>
            </div>
          ) : statistics ? (
            <>
              <div className="flex justify-between text-xs">
                <span className="text-muted-foreground">Storage Used</span>
                <span className="text-sidebar-foreground font-medium">
                  {formatBytes(statistics.total_storage)} / {formatBytes(statistics.storage_quota)}
                </span>
              </div>
              <div className="w-full bg-muted rounded-full h-2">
                <div 
                  className="bg-sidebar-primary h-2 rounded-full transition-all duration-300" 
                  style={{ 
                    width: `${Math.min((statistics.total_storage / statistics.storage_quota) * 100, 100)}%` 
                  }}
                ></div>
              </div>
              <div className="text-xs text-muted-foreground space-y-1">
                <p>{formatBytes(statistics.storage_quota - statistics.total_storage)} remaining</p>
                {statistics.savings.bytes > 0 && (
                  <p className="text-primary">
                   Saved {formatBytes(statistics.savings.bytes)} ({statistics.savings.percentage.toFixed(1)}%)
                  </p>
                )}
              </div>
            </>
          ) : (
            <div className="text-xs text-muted-foreground">
              <p>Unable to load storage info</p>
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

export default DashboardSidebar
