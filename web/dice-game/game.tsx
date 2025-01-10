'use client'

import { useState, useEffect, RefObject } from 'react'
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { ScrollArea } from "@/components/ui/scroll-area"
import Dice from './dice'
import { DebugMenuRef } from './debug-menu'
import { ArrowUp, ArrowDown, Dice1, Dice2, Dice3, Dice4, Dice5, Dice6 } from 'lucide-react'

interface GameProps {
  onEndGame: (score: number) => void
  socket: WebSocket | null
  debugMenuRef: RefObject<DebugMenuRef>
  clientId: string | null
  walletBalance: number
}

export default function Game({ onEndGame, socket, debugMenuRef, clientId, walletBalance }: GameProps) {
  const [points, setPoints] = useState(walletBalance)
  const [bet, setBet] = useState(10)
  const [currentBet, setCurrentBet] = useState(10)
  const [diceValue, setDiceValue] = useState(1)
  const [rolling, setRolling] = useState(false)
  const [history, setHistory] = useState<Array<{ roll: number; result: string; points: number }>>([])

  useEffect(() => {
    if (socket) {
      socket.onmessage = (event) => {
        const data = JSON.parse(event.data)
        debugMenuRef.current?.addReceivedMessage(event.data)

        if (data.kind === 'ROLL') {
          setDiceValue(data.roll)
          setRolling(false)
          updateGameState(data, currentBet)
        } else if (data.kind === 'WALLET') {
          setPoints(data.wallet)
        } else if (data.kind === 'ENDPLAY') {
          setPoints(data.wallet)
          onEndGame(data.result)
        } else if (data.kind === 'ERROR') {
          alert(data.message)
        }
      }
    }
  }, [socket, currentBet])

  const placeBet = (choice: 'ODD' | 'EVEN') => {
    if (bet > points || !socket || !clientId) return

    const message = JSON.stringify({
      kind: 'PLAY',
      clientId: clientId,
      bet: bet,
      choice: choice
    })
    socket.send(message)
    debugMenuRef.current?.addSentMessage(message)

    setCurrentBet(bet)
    setRolling(true)
  }

  const updateGameState = (data: { roll: number; result: string }, bet: number) => {
    const profit = data.result === 'WIN' ? bet : -bet
    setPoints(prevPoints => prevPoints + profit)

    const newHistoryEntry = {
      roll: data.roll,
      result: data.result,
      points: Math.abs(profit)
    }
    setHistory(prev => [newHistoryEntry, ...prev.slice(0, 19)])
  }

  const finishGame = () => {
    if (socket && clientId) {
      const message = JSON.stringify({ kind: 'ENDPLAY', clientId: clientId })
      socket.send(message)
      debugMenuRef.current?.addSentMessage(message)

      const finalScore = points - walletBalance
      onEndGame(finalScore)
    }
  }

  const multiplyBet = (multiplier: number) => {
    setBet(prevBet => {
      const newBet = Math.min(prevBet * multiplier, points)
      return newBet
    })
  }

  const getDiceIcon = (value: number) => {
    const icons = [Dice1, Dice2, Dice3, Dice4, Dice5, Dice6]
    const DiceIcon = icons[value - 1] || Dice1
    return <DiceIcon className="w-6 h-6" />
  }

  return (
    <div className="p-4 h-full" style={{ height: '100%' }}>
      <Card className="w-full mx-auto flex flex-col h-full">
        <CardHeader>
          <CardTitle className="text-3xl text-center">Dice Game</CardTitle>
        </CardHeader>
        <CardContent className="flex-grow flex flex-col overflow-hidden">
          <div className="flex justify-between items-center mb-4">
            <div className="text-3xl font-bold">Points: {points}</div>
            <Dice value={diceValue} rolling={rolling} />
          </div>
          <div className="space-y-4 mb-4">
            <div className="flex items-center gap-2">
              <Input
                type="number"
                value={bet}
                onChange={(e) => {
                  const newBet = Math.max(1, Math.min(parseInt(e.target.value) || 0, points))
                  setBet(newBet)
                }}
                className="w-24"
              />
              <Button onClick={() => multiplyBet(2)} variant="outline" size="sm">x2</Button>
              <Button onClick={() => multiplyBet(5)} variant="outline" size="sm">x5</Button>
              <Button onClick={() => multiplyBet(10)} variant="outline" size="sm">x10</Button>
            </div>
            <div className="flex gap-2">
              <Button onClick={() => placeBet('ODD')} disabled={rolling} className="flex-1">Bet Odd</Button>
              <Button onClick={() => placeBet('EVEN')} disabled={rolling} className="flex-1">Bet Even</Button>
            </div>
          </div>
          <div className='flex flex-col flex-grow'>

            <div className="mt-6 flex-grow flex flex-col">
              <h3 className="font-bold mb-2">Roll History:</h3>
              <ScrollArea className="h-[200px] w-full rounded-md border flex-grow">
                <ul className="space-y-2 p-4">
                  {history.map((entry, index) => (
                    <li
                      key={index}
                      className={`flex items-center gap-2 p-3 rounded-lg ${entry.result === 'WIN'
                        ? 'bg-green-100 dark:bg-green-900'
                        : 'bg-red-100 dark:bg-red-900'
                        }`}
                    >
                      <div className="w-8 h-8 flex items-center justify-center bg-background rounded-full">
                        {getDiceIcon(entry.roll)}
                      </div>
                      <div className="flex-1 flex items-center gap-2">
                        <span className="font-medium">{entry.roll}</span>
                        <span className={`font-medium ${entry.result === 'WIN'
                          ? 'text-green-700 dark:text-green-300'
                          : 'text-red-700 dark:text-red-300'
                          }`}>
                          {entry.result === 'WIN' ? (
                            <ArrowUp className="w-4 h-4 inline" />
                          ) : (
                            <ArrowDown className="w-4 h-4 inline" />
                          )}
                          {entry.points}
                        </span>
                      </div>
                    </li>
                  ))}
                </ul>
              </ScrollArea>
            </div>

            <Button onClick={finishGame} variant="outline" className="w-full mt-4">
              Finish Game
            </Button>


          </div>
        </CardContent>
      </Card>
    </div>
  )
}

