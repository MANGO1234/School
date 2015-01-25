import qualified Data.Map as M
import Data.List
import Data.Maybe

type Player = Char         -- should only be 'W', 'B'
type Position = (Int, Int) -- naming messed up a bit (x, y), x is the row and y is the col
type Board = M.Map Position Player

-- haskell rogram to play Crusher using max-min search

testBoard0 = toBoard 3 "-------------BB-BBB"
testBoard1 = toBoard 3 "WWW-WW-------BB-BBB"
testBoard2 = toBoard 3 "WWW-----W-B-----BBB"

-- moveLeftTop testBoard2 3 'W' (3, 2)

crusher_u4o8 :: [String] -> Player -> Int -> Int -> [String]
crusher_u4o8 [] _ _ _ = []
crusher_u4o8 xs player depth n =
	let history = map (toBoard n) xs
	    board = head history
	    newMove = snd $ crusher_helper_u4o8 player depth n history board
	in (fromBoard newMove n):xs

crusher_helper_u4o8 :: Player -> Int -> Int -> [Board] -> Board -> (Int, Board)
crusher_helper_u4o8 _ 0 _ _ b = (score_u4o8 b, b)
crusher_helper_u4o8 player depth n history b =
	let nextStates = next_states_u4o8 b n player
	    statesBestMove = map (crusher_helper_u4o8 (switch_player player) (depth - 1) n (b:history)) nextStates
	    maximum_move = foldl1' (\a@(x, _) b@(y, _) -> if x > y then a else b) statesBestMove
	    minimum_move = foldl1' (\a@(x, _) b@(y, _) -> if x < y then a else b) statesBestMove
	in if (null nextStates)
	   then (score_u4o8 b, b)
	   else case player of
	        'W' -> maximum_move
	        'B' -> minimum_move

switch_player :: Player -> Player
switch_player 'W' = 'B'
switch_player 'B' = 'W'

next_states_u4o8 :: Board -> Int -> Player -> [Board]
next_states_u4o8 b n player
	| win b = []
	| otherwise = concat [moveLeft b n player,
	                      moveRight b n player,
	                      moveLeftTop b n player,
	                      moveLeftBot b n player,
	                      moveRightTop b n player,
	                      moveRightBot b n player,
	                      jumpLeft b n player,
	                      jumpRight b n player,
	                      jumpLeftTop b n player,
	                      jumpLeftBot b n player,
	                      jumpRightTop b n player,
	                      jumpRightBot b n player]

-- temp
score_u4o8 :: Board -> Int
score_u4o8 b = case (countPlayer b) of
	(0, _) -> -1000000
	(_, 0) -> 10000000
	(x, y) -> x - y

win :: Board -> Bool
win b = let (x, y) = countPlayer b in x == 0 || y == 0

toBoard :: Int -> String -> Board
toBoard n xs = M.fromList pieces
	where tempBoard = to2dList n xs
	      pieces = filter (\(_, ch) -> ch /= '-') tempBoard -- find pieces

-- take a string to a 2 dimensional list version with axial coordinate
to2dList :: Int -> String -> [(Position, Char)]
to2dList n xs = concat $ firstPart xs n (-n+1) 0
	where firstPart xs rowLen rowNo rowStart
	          | rowNo == 0      = newRow:secondPart rest (rowLen - 1) (rowNo + 1) rowStart
	          | otherwise       = newRow:firstPart rest (rowLen + 1) (rowNo + 1) (rowStart - 1)
	          where (row, rest) = splitAt rowLen xs
	                newRow = axialCoordinate row rowLen rowNo rowStart
	      secondPart xs rowLen rowNo rowStart
	          | rowNo == n - 1  = newRow:[]
	          | otherwise       = newRow:secondPart rest (rowLen - 1) (rowNo + 1) rowStart
	          where (row, rest) = splitAt rowLen xs
	                newRow = axialCoordinate row rowLen rowNo rowStart
	      axialCoordinate row rowLen rowNo rowStart = 
	        map (\(y, piece) -> ((rowNo, y), piece)) (zip [rowStart..] row)

-- take a board and convert to string
fromBoard :: Board -> Int -> String
fromBoard b n = concat $ foldl' (replace2d n) (generateEmpty2d n) (M.toList b)

generateEmpty2d :: Int -> [[Char]]
generateEmpty2d n = firstPart n
	where firstPart rowLen
	          | rowLen == 2*n - 1 = newRow:secondPart (rowLen - 1)
	          | otherwise         = newRow:firstPart (rowLen + 1)
	          where newRow = replicate rowLen '-'
	      secondPart rowLen
	          | rowLen == n = newRow:[]
	          | otherwise   = newRow:secondPart (rowLen - 1)
	          where newRow = replicate rowLen '-'

replace2d :: Int -> [[Char]] -> (Position, Player) -> [[Char]]
replace2d n board2d (position, piece) = newBoard
	where (x, y) = coordAxialTo2d n position
	      newRow = replace piece y (board2d !! x)
	      newBoard = replace newRow x board2d

coordAxialTo2d :: Int -> Position -> Position
coordAxialTo2d n (x, y) = (newx, newy)
	where newx = x + n - 1
	      newy = if (x <= 0) then y + n - 1 + x else y + n - 1

-- ************ movement/jumps *********************
-- given a board, the size of the board, the player, and a function that move a piece from a position to another
-- find all the pieces the player control, try applying the function to each piece and if it's a valid move
-- add it to the new candidate boards
moveNewCoord :: Board -> Int -> Player -> (Position -> Position) -> [Board]
moveNewCoord b n player transform = M.foldrWithKey tryNextPiece [] b
	where tryNextPiece coord piece boards
	        | piece == player = if (withinBoard n newCoord) && isNothing (M.lookup newCoord b)
	                            then newBoard:boards
	                            else boards
	        | otherwise = boards
	        where newCoord = transform coord
	              newBoard = M.insert newCoord player $ M.delete coord b

moveLeft :: Board -> Int -> Player -> [Board]
moveLeft b n player = moveNewCoord b n player (\(x, y) -> (x, y - 1))

moveRight :: Board -> Int -> Player -> [Board]
moveRight b n player = moveNewCoord b n player (\(x, y) -> (x, y + 1))

moveLeftTop :: Board -> Int -> Player -> [Board]
moveLeftTop b n player = moveNewCoord b n player (\(x, y) -> (x - 1, y))

moveLeftBot :: Board -> Int -> Player -> [Board]
moveLeftBot b n player = moveNewCoord b n player (\(x, y) -> (x + 1, y - 1))

moveRightTop :: Board -> Int -> Player -> [Board]
moveRightTop b n player = moveNewCoord b n player (\(x, y) -> (x - 1, y + 1))

moveRightBot :: Board -> Int -> Player -> [Board]
moveRightBot b n player = moveNewCoord b n player (\(x, y) -> (x + 1, y + 1))

jumpLeft :: Board -> Int -> Player -> [Board]
jumpLeft b n player = []

jumpRight :: Board -> Int -> Player -> [Board]
jumpRight b n player = []

jumpLeftTop :: Board -> Int -> Player -> [Board]
jumpLeftTop b n player = []

jumpLeftBot :: Board -> Int -> Player -> [Board]
jumpLeftBot b n player = []

jumpRightTop :: Board -> Int -> Player -> [Board]
jumpRightTop b n player = []

jumpRightBot :: Board -> Int -> Player -> [Board]
jumpRightBot b n player = []

withinBoard :: Int -> Position -> Bool
withinBoard n (x, y)
	| x < -n'   = False
	| x > n'    = False
	| otherwise = if (x <= 0)
	              then (-n' - x) <= y && y <= n'
	              else -n' <= y && y <= n' - x
	where n' = n - 1

-- count number of players, first element of tuple is number of 'W', second 'B'
countPlayer :: Board -> (Int, Int)
countPlayer b = foldl' count (0, 0) (M.elems b)
	where count count@(x, y) c
	        | c == 'W'  = (x+1, y)
	        | c == 'B'  = (x, y+1)
	        | otherwise = count

-- replace an element at the nth position of a list with a new value, O(n)
replace :: a -> Int -> [a] -> [a]
replace _ _ []     = []
replace v 0 (_:xs) = v:xs
replace v n (x:xs) = x:replace v (n - 1) xs