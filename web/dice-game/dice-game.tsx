'use client'

import { useState, useEffect, useRef } from 'react'
import Game from './game'
import LandingPage from './landing-page'
import DebugMenu, { DebugMenuRef } from './debug-menu'
import { Button } from "@/components/ui/button"
import { Bug } from 'lucide-react'

export default function DiceGame() {
  const [isConnected, setIsConnected] = useState(false)
  const [gameStarted, setGameStarted] = useState(false);
  const [finalScore, setFinalScore] = useState<number | null>(null);
  const [socket, setSocket] = useState<WebSocket | null>(null);
  const debugMenuRef = useRef<DebugMenuRef>(null);
  const clientId = useRef<string | null>(null);

  useEffect(() => {
    console.log('Connecting to WebSocket server...');

    const ws = new WebSocket('ws://localhost:8080/');

    ws.onopen = () => setIsConnected(true);
    ws.onclose = () => setIsConnected(false);
    ws.onerror = (err) => console.log('WebSocket error:', err);

    setSocket(ws);
    clientId.current = crypto.randomUUID();
    return () => {
      if (ws.readyState === WebSocket.OPEN) ws.close()
    };
  }, []);


  const startGame = () => {
    const message = JSON.stringify({ kind: 'WALLET', clientId: clientId.current });
    socket?.send(message);
    debugMenuRef.current?.addSentMessage(message);

    setGameStarted(true);
    setFinalScore(null);
  };

  const endGame = (score: number) => {
    if (socket) {
      const message = JSON.stringify({ type: 'ENDPLAY', clientId: clientId.current });
      socket.send(message);
      debugMenuRef.current?.addSentMessage(message);
    }

    setGameStarted(false);
    setFinalScore(score);
  };

  const toggleDebugMenu = () => {
    debugMenuRef.current?.toggleMenu();
  };

  return (
    <div className="min-h-screen bg-gray-100 relative">
      <div className="mx-auto">
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
          <Game onEndGame={endGame} socket={socket} debugMenuRef={debugMenuRef} />
        ) : (
          <LandingPage onStartGame={startGame} finalScore={finalScore} />
        )}
      </div>
    </div>
  )
}

