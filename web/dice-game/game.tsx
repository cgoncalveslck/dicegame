'use client'

import { useState, useEffect, RefObject } from 'react'
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import Dice from './dice'
import { DebugMenuRef } from './debug-menu'

interface GameProps {
  onEndGame: (score: number) => void
  socket: WebSocket | null
  debugMenuRef: RefObject<DebugMenuRef>
}

export default function Game({ onEndGame, socket, debugMenuRef }: GameProps) {
  const [points, setPoints] = useState(1000)
  const [bet, setBet] = useState(10)
  const [diceValue, setDiceValue] = useState(1)
  const [rolling, setRolling] = useState(false)
  const [history, setHistory] = useState<string[]>([])

  useEffect(() => {
    if (socket) {
      socket.onmessage = (event) => {
        const data = JSON.parse(event.data)
        if (data.type === 'ROLL_RESULT') {
          setDiceValue(data.value)
          setRolling(false)
          updateGameState(data)
        }
      }
    }
  }, [socket])

  const rollDice = () => {
    setRolling(true)

    setTimeout(() => {
      const newValue = Math.floor(Math.random() * 6) + 1
      setDiceValue(newValue)
      setRolling(false)
    }, 1000)
  }

  const placeBet = (choice: 'odd' | 'even') => {
    const choiceInt = choice === 'odd' ? 1 : 2
    if (bet > points) return
    if (socket) {
      const message = JSON.stringify({ kind: 'PLAY', data: { bet, choice: choiceInt } })
      socket.send(message)
      debugMenuRef.current?.addSentMessage(message)
    }
    rollDice()
  }

  const updateGameState = (data: { value: number; won: boolean; pointChange: number }) => {
    setPoints(prevPoints => prevPoints + data.pointChange)
    setHistory(prev => [`${data.value} (${data.won ? 'Won' : 'Lost'} ${Math.abs(data.pointChange)})`, ...prev.slice(0, 4)])
  }

  const finishGame = () => {
    const finalScore = points - 1000 // Calculate net profit
    onEndGame(finalScore)
  }

  const multiplyBet = (multiplier: number) => {
    setBet(prevBet => Math.min(prevBet * multiplier, points))
  }

  return (
    <div className="p-4">
      <Card className="w-full max-w-2xl mx-auto">
        <CardHeader>
          <CardTitle className="text-3xl text-center">Dice Game</CardTitle>
        </CardHeader>
        <CardContent>
          <div className="flex justify-between items-center mb-6">
            <div className="text-3xl font-bold">Points: {points}</div>
            <Dice value={diceValue} rolling={rolling} />
          </div>
          <div className="space-y-4">
            <div className="flex items-center gap-2">
              <Input
                type="number"
                value={bet}
                onChange={(e) => setBet(Math.max(1, Math.min(parseInt(e.target.value) || 0, points)))}
                className="w-24"
              />
              <Button onClick={() => multiplyBet(2)} variant="outline" size="sm">x2</Button>
              <Button onClick={() => multiplyBet(5)} variant="outline" size="sm">x5</Button>
              <Button onClick={() => multiplyBet(10)} variant="outline" size="sm">x10</Button>
            </div>
            <div className="flex gap-2">
              <Button onClick={() => placeBet('odd')} disabled={rolling} className="flex-1">Bet Odd</Button>
              <Button onClick={() => placeBet('even')} disabled={rolling} className="flex-1">Bet Even</Button>
            </div>
          </div>
          <div className="mt-6">
            <h3 className="font-bold mb-2">Roll History:</h3>
            <ul className="space-y-1">
              {history.map((roll, index) => (
                <li key={index} className="text-sm">{roll}</li>
              ))}
            </ul>
          </div>
          <Button onClick={finishGame} variant="outline" className="w-full mt-6">
            Finish Game
          </Button>
        </CardContent>
      </Card>
    </div>
  )
}

