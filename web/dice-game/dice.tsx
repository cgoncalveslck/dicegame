import { useState, useEffect } from 'react'

interface DiceProps {
  value: number
  rolling: boolean
}

export default function Dice({ value, rolling }: DiceProps) {
  const [displayValue, setDisplayValue] = useState(value)

  useEffect(() => {
    if (rolling) {
      const interval = setInterval(() => {
        setDisplayValue(Math.floor(Math.random() * 6) + 1)
      }, 100)
      return () => clearInterval(interval)
    } else {
      setDisplayValue(value)
    }
  }, [rolling, value])

  return (
    <div className={`w-24 h-24 bg-white rounded-lg shadow-md flex items-center justify-center text-6xl font-bold ${rolling ? 'animate-bounce' : ''}`}>
      {displayValue}
    </div>
  )
}

