import { Button } from "@/components/ui/button"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"

interface LandingPageProps {
  onStartGame: () => void
  finalScore: number | null
}

export default function LandingPage({ onStartGame, finalScore }: LandingPageProps) {
  return (
    <div className="container mx-auto p-4 h-screen flex items-center justify-center">
      <Card className="w-full max-w-md">
        <CardHeader>
          <CardTitle className="text-3xl text-center">Welcome to the Dice Game!</CardTitle>
        </CardHeader>
        <CardContent className="text-center">
          <p className="mb-6">Test your luck by betting on odd or even dice rolls.</p>
          {finalScore !== null && (
            <p className="mb-6 text-xl font-bold">
              Your session profit: {finalScore > 0 ? '+' : ''}{finalScore} points
            </p>
          )}
          <Button onClick={onStartGame} size="lg">
            {finalScore !== null ? 'Play Again' : 'Start Game'}
          </Button>
        </CardContent>
      </Card>
    </div>
  )
}
