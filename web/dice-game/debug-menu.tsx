import { useState, useEffect, forwardRef, useImperativeHandle } from 'react'
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { ScrollArea } from "@/components/ui/scroll-area"
import { Button } from "@/components/ui/button"
import { X } from 'lucide-react'

interface DebugMenuProps {
  socket: WebSocket | null
}

export interface DebugMenuRef {
  addSentMessage: (message: string) => void
  addReceivedMessage: (message: string) => void

  toggleMenu: () => void
}

const DebugMenu = forwardRef<DebugMenuRef, DebugMenuProps>(({ socket }, ref) => {
  const [isConnected, setIsConnected] = useState(false)
  const [messages, setMessages] = useState<string[]>([])
  const [isVisible, setIsVisible] = useState(false)

  useEffect(() => {
    if (socket) {
      const handleOpen = () => setIsConnected(true)
      const handleClose = () => setIsConnected(false)
      const handleMessage = (event: MessageEvent) => {
        setMessages(prev => [`Received: ${event.data}`, ...prev])
      }

      socket.addEventListener('open', handleOpen)
      socket.addEventListener('close', handleClose)
      socket.addEventListener('message', handleMessage)

      return () => {
        socket.removeEventListener('open', handleOpen)
        socket.removeEventListener('close', handleClose)
        socket.removeEventListener('message', handleMessage)
      }
    }
  }, [socket])

  const addSentMessage = (message: string) => {
    setMessages(prev => [`Sent: ${message}`, ...prev])
  }

  const addReceivedMessage = (message: string) => {
    setMessages(prev => [`Received: ${message}`, ...prev])
  }

  const toggleMenu = () => {
    setIsVisible(prev => !prev)
  }

  useImperativeHandle(ref, () => ({
    addSentMessage,
    addReceivedMessage,
    toggleMenu
  }))

  if (!isVisible) return null

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <Card className="w-full max-w-2xl max-h-[80vh] flex flex-col">
        <CardHeader className="flex flex-row items-center">
          <CardTitle className="text-xl flex-grow">Debug Menu</CardTitle>
          <Button variant="ghost" size="icon" onClick={toggleMenu}>
            <X className="h-4 w-4" />
          </Button>
        </CardHeader>
        <CardContent className="flex-grow overflow-hidden">
          <div className="mb-4">
            <strong>WebSocket Status:</strong> {isConnected ? 'Connected' : 'Disconnected'}
          </div>
          <div className="h-full">
            <strong>Message Log:</strong>
            <ScrollArea className="h-[calc(100%-2rem)] w-full border rounded-md p-2">
              {messages.map((message, index) => (
                <div key={index} className="text-sm">
                  {message}
                </div>
              ))}
            </ScrollArea>
          </div>
        </CardContent>
      </Card>
    </div>
  )
})

DebugMenu.displayName = 'DebugMenu'

export default DebugMenu

