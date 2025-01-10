'use client'

import { useState, useEffect, useRef } from 'react'
import Game from './game'
import LandingPage from './landing-page'
import DebugMenu, { DebugMenuRef } from './debug-menu'
import { Button } from "@/components/ui/button"
import { Bug } from 'lucide-react'

export default function DiceGame() {
  const [isConnected, setIsConnected] = useState(false)
  const [gameStarted, setGameStarted] = useState(false)
  const [finalScore, setFinalScore] = useState<number | null>(null)
  const [socket, setSocket] = useState<WebSocket | null>(null)
  const [walletBalance, setWalletBalance] = useState<number | null>(null)
  const debugMenuRef = useRef<DebugMenuRef>(null)
  const clientId = useRef<string | null>(null)

  useEffect(() => {
    const ws = new WebSocket('ws://localhost:81818/')

    const handleOpen = () => {
      setIsConnected(true)
      const authMessage = JSON.stringify({ kind: 'AUTH' })
      ws.send(authMessage)
      debugMenuRef.current?.addSentMessage(authMessage)
    }

    const handleClose = () => setIsConnected(false)

    const handleMessage = (event: MessageEvent) => {
      const data = JSON.parse(event.data)

      if (data.kind === 'AUTH') {
        clientId.current = data.clientId
      } else if (data.kind === 'WALLET') {
        setWalletBalance(data.wallet)
      }
    }

    ws.onopen = handleOpen
    ws.onclose = handleClose
    ws.onmessage = handleMessage

    setSocket(ws)

    return () => {
      ws.removeEventListener('open', handleOpen)
      ws.removeEventListener('close', handleClose)
      ws.removeEventListener('message', handleMessage)
      if (ws.readyState === WebSocket.OPEN) ws.close()
    }
  }, [])

  const startGame = () => {
    if (socket && clientId.current) {
      const walletMessage = JSON.stringify({ kind: 'WALLET', clientId: clientId.current })
      socket.send(walletMessage)
      debugMenuRef.current?.addSentMessage(walletMessage)

      const startMessage = JSON.stringify({ kind: 'STARTPLAY', clientId: clientId.current })
      socket.send(startMessage)
      debugMenuRef.current?.addSentMessage(startMessage)

      setGameStarted(true)
      setFinalScore(null)
    }
  }
  const endGame = (score: number) => {
    setGameStarted(false)
    setFinalScore(score)
  }


  const toggleDebugMenu = () => {
    debugMenuRef.current?.toggleMenu()
  }

  return (
    <div className="min-h-screen bg-gray-100 relative">
      <div className="mx-auto" style={{ height: '100vh' }}>
        <Button
          onClick={toggleDebugMenu}
          className="fixed bottom-4 right-4 z-10 rounded-full"
          variant="outline"
          size="icon"
          style={{ backgroundColor: isConnected ? '#43ff64d9' : '#ff0000b3' }}
        >
          <Bug className="h-4 w-4" />
        </Button>
        <DebugMenu socket={socket} ref={debugMenuRef} />
        {gameStarted ? (
          <Game
            onEndGame={endGame}
            socket={socket}
            debugMenuRef={debugMenuRef}
            clientId={clientId.current}
            walletBalance={walletBalance || 0}
          />
        ) : (
          <LandingPage onStartGame={startGame} finalScore={finalScore} />
        )}
      </div>
    </div>
  )
}
